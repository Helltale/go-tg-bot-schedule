package handlecommand

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	pbac "tgclient/proto/adress-contact"
	pba "tgclient/proto/auth"
	pbs "tgclient/proto/schedule"
	pbt "tgclient/proto/teacher"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthState int

const (
	StateStart AuthState = iota
	StateAwaitingName
	StateAwaitingGroup
	StateAuthorized //главное меню
	StateScheduleMenu
	StateToMainMenu                    //для возврата в главное меню
	StateFindTeacherMenu               // меню для поиска преподавателя
	StateFindTeacherAwaitingFIO        //поиск препода по фио, ожидание ввода
	StateFindTeacherAwaitingDepartment //поиск препода по кафедре, ожидание ввода
	StateFindTeacherAwaitingSubject    //поиск препода по предмету, ожидание ввода
	StateAdressContactMenu             //меню для адресов и контактов
	StateDocumentMenu                  //меню для документов
	StateDocumentGroup1Menu            //меню 1 группы
	StateReadyForDownloadDocument      //готовность для скачивания файла
)

type AuthContext struct {
	State          AuthState
	UserID         int64
	ProfileName    string
	LastMessageID  int64
	LastMessageIDs []int64
	LastPlaceName  string
	SelectedPlace  *pbac.Place
}

type PlaceInfo struct {
	Name              string
	WorkTime          string
	Phone             string
	Email             string
	Address           string
	PlaceAddressPoint string
}

// Маппинг для дней недели
var daysOfWeek = map[time.Weekday]string{
	time.Sunday:    "Воскресенье",
	time.Monday:    "Понедельник",
	time.Tuesday:   "Вторник",
	time.Wednesday: "Среда",
	time.Thursday:  "Четверг",
	time.Friday:    "Пятница",
	time.Saturday:  "Суббота",
}

func New(bot *telego.Bot, updates <-chan telego.Update, options ...th.BotHandlerOption) {
	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Fatalf("Ошибка при создании обработчика бота: %v", err)
	}
	defer bh.Stop()

	authContext := &AuthContext{State: StateStart}

	//start
	bh.Handle(func(b *telego.Bot, update telego.Update) {
		authContext.UserID = update.Message.Chat.ID
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// Проверяем, зарегистрирован ли пользователь
		resp, err := checkUser(ctx, authContext.UserID)
		if err != nil {
			sendMessage(b, update.Message.Chat.ID, "Ошибка при проверке пользователя. Пожалуйста, попробуйте снова.", authContext)
			return
		}

		if resp.Exists {
			authContext.ProfileName = resp.ProfileName
			authContext.State = StateAuthorized
			// sendfirstMessage(b, update.Message.Chat.ID, "Вы успешно авторизованы!")
			sendMainMenu(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID) // Открываем главное меню
		} else {
			authContext.State = StateAwaitingName
			sendMessageWithoutDelete(b, update.Message.Chat.ID, "Добро пожаловать!\nПожалуйста, введите ваше ФИО:", authContext)
		}
	}, th.CommandEqual("start"))

	bh.Handle(func(b *telego.Bot, update telego.Update) {
		if update.Message != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			switch authContext.State {
			case StateStart:
				authContext.UserID = update.Message.Chat.ID
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				resp, err := checkUser(ctx, authContext.UserID)
				if err != nil {
					sendMessage(b, update.Message.Chat.ID, "Ошибка при проверке пользователя. Пожалуйста, попробуйте снова.", authContext)
					return
				}

				if resp.Exists {
					authContext.ProfileName = resp.ProfileName
					authContext.State = StateAuthorized
					sendMessage(b, update.Message.Chat.ID, "Вы успешно авторизованы!", authContext)

					sendScheduleMenu(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID)
				} else {
					sendMessage(b, update.Message.Chat.ID, "Пользователь не найден. Пожалуйста, введите ваше имя:", authContext)
					authContext.State = StateAwaitingName
				}
			case StateAwaitingName:
				authContext.ProfileName = update.Message.Text
				sendGroupSelection(b, update.Message.Chat.ID, authContext)
				authContext.State = StateAwaitingGroup

			case StateFindTeacherAwaitingFIO:
				teachers, err := findTeachersByFIO(ctx, update.Message.Text)
				if err != nil {
					sendMessage(b, update.Message.Chat.ID, fmt.Sprintf("Ошибка при поиске преподавателей: %v", err), authContext)
					return
				}
				sendTeachersInfo(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID, teachers)
				authContext.State = StateToMainMenu // Вернуться

			case StateFindTeacherAwaitingDepartment:
				teachers, err := findTeachersByDepartment(ctx, update.Message.Text)
				if err != nil {
					sendMessage(b, update.Message.Chat.ID, fmt.Sprintf("Ошибка при поиске преподавателей: %v", err), authContext)
					return
				}
				sendTeachersInfo(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID, teachers)
				authContext.State = StateToMainMenu // Вернуться к авторизованному состоянию

			case StateFindTeacherAwaitingSubject:
				teachers, err := findTeachersBySubject(ctx, update.Message.Text)
				if err != nil {
					sendMessage(b, update.Message.Chat.ID, fmt.Sprintf("Ошибка при поиске преподавателей: %v", err), authContext)
					return
				}
				sendTeachersInfo(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID, teachers)
				authContext.State = StateToMainMenu // Вернуться к авторизованному состоянию

			}

		} else if update.CallbackQuery != nil {
			if authContext.State == StateAwaitingGroup {
				group := update.CallbackQuery.Data

				sendMessage(b, update.CallbackQuery.From.ID, fmt.Sprintf("Вы выбрали группу: %s", group), authContext)

				// Регистрация пользователя
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := registerUser(ctx, authContext.UserID, authContext.ProfileName, group) // Передаем группу
				if err != nil {
					sendMessage(b, update.CallbackQuery.From.ID, "Ошибка при регистрации. Пожалуйста, попробуйте снова.", authContext)
					return
				}
				authContext.State = StateAuthorized

				sendMessageWithoutDelete(b, update.CallbackQuery.From.ID, "Вы успешно зарегистрированы!", authContext)
				sendMainMenu(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID)
			} else if authContext.State == StateAuthorized {
				// Обработка нажатий на кнопки расписания

				if update.CallbackQuery.Data == "schedule_today" {
					sendSchedule(b, update.CallbackQuery.From.ID, "today", authContext.ProfileName, authContext)

				} else if update.CallbackQuery.Data == "schedule_week" {
					sendSchedule(b, update.CallbackQuery.From.ID, "week", authContext.ProfileName, authContext)

				}
			} else if authContext.State == StateToMainMenu {
				// Обработка нажатий на кнопки расписания

				if update.CallbackQuery.Data == "main_menu" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.Message.Chat.ID)
				}
			}
		}
	}, th.AnyMessage())

	bh.Handle(func(b *telego.Bot, update telego.Update) {
		if update.CallbackQuery != nil {

			if authContext.State == StateAwaitingGroup {
				group := update.CallbackQuery.Data
				sendMessageWithoutDelete(b, update.CallbackQuery.From.ID, fmt.Sprintf("Вы выбрали группу: %s", group), authContext)

				// Регистрация пользователя
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := registerUser(ctx, authContext.UserID, authContext.ProfileName, group) // Передаем группу
				if err != nil {
					sendMessageWithoutDelete(b, update.CallbackQuery.From.ID, "Ошибка при регистрации. Пожалуйста, попробуйте снова.", authContext)
					return
				}

				authContext.State = StateAuthorized
				sendMessageWithoutDelete(b, update.CallbackQuery.From.ID, "Вы успешно зарегистрированы и авторизованы!", authContext)
				message := fmt.Sprintf(`Чтобы быть в курсе последних новостей, скорее перейдите по ссылке, если ещё не подписаны: https://t.me/pmishSamSMU 
	Буду рад ответить на любые ваши вопросы, %s!`, authContext.ProfileName)

				sendMessageWithoutDelete(b, update.CallbackQuery.From.ID, message, authContext)
				sendMainMenuWithoutDelete(b, update.CallbackQuery.From.ID, authContext)

			}

			if authContext.State == StateAuthorized {

				switch update.CallbackQuery.Data {
				case "schedule":

					authContext.State = StateScheduleMenu

					//мб тут проблема? некорректно передаются параметры
					// sendMessage(b, update.CallbackQuery.From.ID, fmt.Sprintf("userid %d", update.CallbackQuery.From.ID))
					// sendMessage(b, update.CallbackQuery.From.ID, fmt.Sprintf("authcontext %d", authContext.LastMessageID))
					// sendMessage(b, update.CallbackQuery.From.ID, fmt.Sprintf("chatid %d", firstMessageID))

					sendScheduleMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
				case "teachers":
					// Обработка нажатия на "Преподаватели"

					authContext.State = StateFindTeacherMenu
					sendTeachersInfoMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
				case "contacts":
					authContext.State = StateAdressContactMenu
					sendAdressContactMenu(bot, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
				case "documents":
					// Обработка нажатия на "Шаблоны/бланки документов"
					sendDocumentsInfo(b, update.CallbackQuery.From.ID, authContext)
				case "extracurricular":
					// Обработка нажатия на "Внеурочная активная деятельность"
					sendExtracurricularInfo(b, update.CallbackQuery.From.ID, authContext)
				case "ask_question":
					// Обработка нажатия на "Задать вопрос"
					sendQuestionForm(b, update.CallbackQuery.From.ID, authContext)
				case "change_user":
					// Обработка нажатия на "Сменить пользователя"
					authContext.State = StateStart
					sendMessage(b, update.CallbackQuery.From.ID, "Вы вышли из аккаунта. Пожалуйста, введите ваше имя:", authContext)
				}
			}

			if authContext.State == StateScheduleMenu {
				// Обработка нажатий на кнопки расписания
				if update.CallbackQuery.Data == "schedule_today" {
					authContext.State = StateToMainMenu
					getGroupAndSendSchedule(b, update.CallbackQuery.From.ID, "today", authContext.UserID, authContext, update.CallbackQuery.From.ID)

					// toMainMenu(b, update.CallbackQuery.From.ID)
				} else if update.CallbackQuery.Data == "schedule_week" {
					authContext.State = StateToMainMenu
					getGroupAndSendSchedule(b, update.CallbackQuery.From.ID, "week", authContext.UserID, authContext, update.CallbackQuery.From.ID)

					// toMainMenu(b, update.CallbackQuery.From.ID)
				} else if update.CallbackQuery.Data == "schedule_back" { //кнопка назад (к главному меню) из выбора расписания расписания
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized
				}
			}

			if authContext.State == StateToMainMenu {
				if update.CallbackQuery.Data == "main_menu" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized
				}
			}

			if authContext.State == StateFindTeacherMenu {
				// Обработка нажатий на маню посика препода
				if update.CallbackQuery.Data == "teacher_find_back" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized
				} else if update.CallbackQuery.Data == "teacher_find_fio" {
					sendMessage(b, update.CallbackQuery.From.ID, "Введите фио преподавателя:", authContext)
					authContext.State = StateFindTeacherAwaitingFIO
				} else if update.CallbackQuery.Data == "teacher_find_department" {
					sendMessage(b, update.CallbackQuery.From.ID, "Введите кафедру преподавателя:", authContext)
					authContext.State = StateFindTeacherAwaitingDepartment
				} else if update.CallbackQuery.Data == "teacher_find_subject" {
					sendMessage(b, update.CallbackQuery.From.ID, "Введите предмет:", authContext)
					authContext.State = StateFindTeacherAwaitingSubject
				}

			}

			if authContext.State == StateAdressContactMenu {
				if update.CallbackQuery.Data == "adress_contact_back" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized

				} else if update.CallbackQuery.Data == "adress_contact_administrative" {
					place := "Административный корпус"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = StateToMainMenu

				} else if update.CallbackQuery.Data == "adress_contact_study" {
					place := "Учебный корпус"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = StateToMainMenu

				} else if update.CallbackQuery.Data == "adress_contact_living" {
					place := "Общежитие"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = StateToMainMenu

				} else if update.CallbackQuery.Data == "adress_contact_departments" {
					place := "Кафедры"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = StateToMainMenu

				}
			}

			if authContext.State == StateToMainMenu {
				if update.CallbackQuery.Data == "send_location" {

					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()

					if authContext.SelectedPlace != nil {
						// Получаем информацию о месте
						places, err := findAddressByPlaceName(ctx, authContext.SelectedPlace.PlaceName) // Используем сохраненное место
						if err != nil {
							log.Printf("Ошибка при получении адреса: %v", err)
							return
						}

						// Предполагаем, что вы хотите отправить локацию первого места
						if len(places) > 0 {
							latitude := places[0].PlaceAdressPoint.Latitude
							longitude := places[0].PlaceAdressPoint.Longitude
							sendMessageAdressLocation(bot, update.CallbackQuery.From.ID, authContext, latitude, longitude)
						} else {
							log.Println("Нет доступных мест для отправки локации.")
						}
					} else {
						log.Println("Нет выбранного места.")
					}

				}
			}

			if authContext.State == StateDocumentMenu {
				if update.CallbackQuery.Data == "documents_back" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized
				} else if update.CallbackQuery.Data == "document_group_1" {
					sendDocumentsGroup1Menu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateDocumentGroup1Menu
				}
			}

			if authContext.State == StateDocumentGroup1Menu {
				if update.CallbackQuery.Data == "document_group_1_doc1" {
					message := getFileInfo("doc1")
					// sendMessage(b, update.CallbackQuery.From.ID, message, authContext)
					sendFileInfoMessage(b, update.CallbackQuery.From.ID, message, authContext)
					authContext.State = StateReadyForDownloadDocument

				} else if update.CallbackQuery.Data == "document_group_1_doc2" {
					sendMessage(b, update.CallbackQuery.From.ID, "doc2", authContext)
				} else if update.CallbackQuery.Data == "document_group_1_doc3" {
					sendMessage(b, update.CallbackQuery.From.ID, "doc3", authContext)
				} else if update.CallbackQuery.Data == "main_menu" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized
				}
			}

			if authContext.State == StateReadyForDownloadDocument {
				if update.CallbackQuery.Data == "main_menu" {
					sendMainMenu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = StateAuthorized
				} else if update.CallbackQuery.Data == "downloading_file" {
					sendFileInfoDocument(b, update.CallbackQuery.From.ID, authContext)
				}
			}

		}
	}, th.AnyCallbackQuery())

	bh.Start()
}

func getFileInfo(document string) string {
	var message string

	switch document {
	case "doc1":
		fileInfo, err := os.Stat(filepath.Join("d:\\", "projects", "golang", "tg-program", "client-tg", "other", "docs", document+".docx"))
		if err != nil {
			fmt.Println("Ошибка:", err)
			return message
		}
		message += fmt.Sprintf("Название: %s\n", fileInfo.Name())
		message += fmt.Sprintf("Размер: %d б\n", fileInfo.Size())
		message += fmt.Sprintf("Дата обновления: %s\n", fileInfo.ModTime().Format(time.DateOnly))

		return message

	default:
		return message
	}
}

func sendFileInfoMessage(bot *telego.Bot, userID int64, messageIn string, authContext *AuthContext) {

	//очистка предыдущих сообщений
	clearMessages(bot, authContext, userID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Скачать").WithCallbackData("downloading_file"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		messageIn,
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке результата поиска: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendFileInfoDocument(bot *telego.Bot, userID int64, authContext *AuthContext) {

	clearMessages(bot, authContext, userID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
		),
	)

	document := tu.Document(
		tu.ID(userID),
		tu.File(mustOpen2("doc1")),
	).WithReplyMarkup(inlineKeyboard)
	// Sending document
	sentMessage, err := bot.SendDocument(document)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendTeachersInfo(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64, teachers []*pbt.Teacher) {

	//сообщения полученные от сервера
	messages, imgs := TextForTeacherInfo(teachers)

	//очистка предыдущих сообщений
	clearMessages(bot, authContext, chatID)

	//если ответ пуст то выдавать сообщение о том что пусто
	if len(messages) == 0 {
		inlineKeyboard := tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
			),
		)

		message := tu.Message(
			tu.ID(userID),
			"Никого не найдено :с",
		).WithReplyMarkup(inlineKeyboard)

		sentMessage, err := bot.SendMessage(message)
		if err != nil {
			log.Printf("Ошибка при отправке результата поиска: %v", err)
		} else {
			authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		}
		return
	}

	for index, message := range messages {

		inlineKeyboard := tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
			),
		)

		photo := tu.Photo(
			tu.ID(userID),
			tu.File(mustOpen1(imgs[index])),
		).WithReplyMarkup(inlineKeyboard).WithCaption(message)

		sentMessage, err := bot.SendPhoto(photo)
		if err != nil {
			log.Printf("Ошибка при отправке результата поиска: %v", err)
		} else {
			authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		}
	}

}

func mustOpen1(filename string) *os.File {
	// Добавляем .jpg к имени файла
	filePath := filepath.Join("D:\\", "projects", "golang", "tg-program", "client-tg", "other", "imgs", filename+".jpg")

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Ошибка: файл не найден по пути: %s", filePath)
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", filePath, err)
	}
	return file
}

func mustOpen2(filename string) *os.File {
	// Добавляем .jpg к имени файла
	filePath := filepath.Join("D:\\", "projects", "golang", "tg-program", "client-tg", "other", "docs", filename+".docx")

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Ошибка: файл не найден по пути: %s", filePath)
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", filePath, err)
	}
	return file
}

func TextForTeacherInfo(teachers []*pbt.Teacher) ([]string, []string) {
	if len(teachers) == 0 {
		return nil, nil
	}

	var messages []string
	var imgs []string
	for _, teacher := range teachers {
		message := fmt.Sprintf("ФИО: %s\n", teacher.TeacherName)
		message += fmt.Sprintf("Должность: %s\n", teacher.TeacherJob)
		message += fmt.Sprintf("Кафедра: %s\n", teacher.TeacherDepartment)
		message += fmt.Sprintf("Адрес: %s\n", teacher.TeacherAdress)
		message += fmt.Sprintf("Почта: %s\n", teacher.TeacherEmail)

		messages = append(messages, message)
		imgs = append(imgs, teacher.ImageName)
	}
	return messages, imgs
}

func sendMainMenu(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Преподаватели").WithCallbackData("teachers"),
			tu.InlineKeyboardButton("Расписание").WithCallbackData("schedule"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Адреса и контакты").WithCallbackData("contacts"),
			tu.InlineKeyboardButton("Бланки документов").WithCallbackData("documents"),
		),
		// tu.InlineKeyboardRow(
		// 	tu.InlineKeyboardButton("Внеурочная активная деятельность").WithCallbackData("extracurricular"),
		// 	tu.InlineKeyboardButton("Задать вопрос").WithCallbackData("ask_question"),
		// ),
	)

	message := tu.Message(
		tu.ID(userID),
		"Возможно, Ваш вопрос уже представлен в навигационном меню.\nПопробуйте выбрать подходящий вариант:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendDocumentsMenu(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Группа 1").WithCallbackData("document_group_1"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Группа 2").WithCallbackData("document_group_2"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Группа 3").WithCallbackData("document_group_3"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Назад").WithCallbackData("documents_back"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		"Выберите из списка:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendDocumentsGroup1Menu(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 1").WithCallbackData("document_group_1_doc1"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 2").WithCallbackData("document_group_1_doc2"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 3").WithCallbackData("document_group_1_doc3"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Назад").WithCallbackData("main_menu"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		"Выберите из списка:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendMainMenuWithoutDelete(bot *telego.Bot, userID int64, authContext *AuthContext) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Преподаватели").WithCallbackData("teachers"),
			tu.InlineKeyboardButton("Расписание").WithCallbackData("schedule"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Адреса и контакты").WithCallbackData("contacts"),
			tu.InlineKeyboardButton("Шаблоны/бланки документов").WithCallbackData("documents"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Внеурочная активная деятельность").WithCallbackData("extracurricular"),
			tu.InlineKeyboardButton("Задать вопрос").WithCallbackData("ask_question"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		"Возможно, Ваш вопрос уже представлен в навигационном меню.\nПопробуйте выбрать подходящий вариант:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		// authContext.LastMessageID = int64(sentMessage.MessageID)
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		// sendMessage(bot, chatID, fmt.Sprintf("ид последнего сообщения `%d`", authContext.LastMessageID))
	}
}

func toMainMenu2(bot *telego.Bot, userID int64, messageIn string, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		messageIn,
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		authContext.LastMessageID = int64(sentMessage.MessageID)
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		// sendMessage(bot, chatID, fmt.Sprintf("ид последнего сообщения `%d`", authContext.LastMessageID))
	}
}

func sendAdressContactMenu(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Административный корпус").WithCallbackData("adress_contact_administrative"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Учебный корпус").WithCallbackData("adress_contact_study"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Общежитие").WithCallbackData("adress_contact_living"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Кафедры").WithCallbackData("adress_contact_departments"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Назад").WithCallbackData("adress_contact_back"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		"Выберите интересующий вариант:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		// authContext.LastMessageID = int64(sentMessage.MessageID)
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		// sendMessage(bot, chatID, fmt.Sprintf("ид последнего сообщения `%d`", authContext.LastMessageID))
	}
}

func sendScheduleMenu(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("На неделю").WithCallbackData("schedule_week"),
			tu.InlineKeyboardButton("На сегодня").WithCallbackData("schedule_today"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Назад").WithCallbackData("schedule_back"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		"Выберите тип расписания:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		// authContext.LastMessageID = int64(sentMessage.MessageID)
	}
}

func sendGroupSelection(bot *telego.Bot, userID int64, authContext *AuthContext) {
	clearMessages(bot, authContext, userID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	groupsResp, err := getGroups(ctx)
	if err != nil {
		sendMessage(bot, userID, "Ошибка при получении групп. Пожалуйста, попробуйте снова.", authContext)
		return
	}

	var inlineButtons []telego.InlineKeyboardButton
	for _, group := range groupsResp.Groups {
		inlineButtons = append(inlineButtons,
			tu.InlineKeyboardButton(group).WithCallbackData(group),
		)
	}

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(inlineButtons...),
	)

	message := tu.Message(
		tu.ID(userID),
		"Выберите вашу группу:",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке выбора группы: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendSchedule(bot *telego.Bot, userID int64, requestType, groupName string, authContext *AuthContext) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Выполняем gRPC вызов для получения расписания
	scheduleResp, err := getSchedule(ctx, groupName, requestType)
	if err != nil {
		sendMessage(bot, userID, fmt.Sprintf("Ошибка при получении расписания. Пожалуйста, попробуйте снова. %s", err), authContext)
		return
	}

	// Формируем сообщение с расписанием
	if len(scheduleResp.Lessons) == 0 {
		sendMessage(bot, userID, "Расписание пусто.", authContext)
		return
	}

	// Создаем мапу для хранения расписания по дням
	type DaySchedule struct {
		Date    string
		Weekday string
		Lessons []*pbs.Lesson
	}

	schedules := make(map[string]*DaySchedule)

	for _, lesson := range scheduleResp.Lessons {
		// Парсим время начала
		startTime, err := time.Parse(time.RFC3339, lesson.StartTime)
		if err != nil {
			sendMessage(bot, userID, "Ошибка при обработке даты.", authContext)
			return
		}

		// Получаем день недели и дату
		weekday := daysOfWeek[startTime.Weekday()]
		formattedDate := startTime.Format("02.01.2006") // Форматируем дату

		// Проверяем, существует ли уже запись для этой даты
		if daySchedule, exists := schedules[formattedDate]; exists {
			// Если запись существует, добавляем урок
			daySchedule.Lessons = append(daySchedule.Lessons, lesson)
		} else {
			// Если записи нет, создаем новую
			schedules[formattedDate] = &DaySchedule{
				Date:    formattedDate,
				Weekday: weekday,
				Lessons: []*pbs.Lesson{lesson},
			}
		}
	}

	// Сортируем расписание по дате
	var sortedSchedules []*DaySchedule
	for _, daySchedule := range schedules {
		sortedSchedules = append(sortedSchedules, daySchedule)
	}
	sort.Slice(sortedSchedules, func(i, j int) bool {
		return sortedSchedules[i].Date < sortedSchedules[j].Date
	})

	// Формируем сообщение
	var message string
	for _, daySchedule := range sortedSchedules {
		message += fmt.Sprintf("🔴 Расписание на %s (%s):\n", daySchedule.Date, daySchedule.Weekday)
		for _, lesson := range daySchedule.Lessons {
			// Парсим время начала и конца
			startTime, _ := time.Parse(time.RFC3339, lesson.StartTime)
			endTime, _ := time.Parse(time.RFC3339, lesson.EndTime)

			// Форматируем время
			formattedTime := fmt.Sprintf("%s - %s", startTime.Format("15:04"), endTime.Format("15:04"))

			// Проверяем SubjectName и заменяем его при необходимости
			TypeEducation := lesson.TypeEducation
			if TypeEducation == "Практическое занятие" {
				TypeEducation = "Практика"
			}
			if TypeEducation == "Лабораторное занятие" {
				TypeEducation = "Лабораторная"
			}

			message += fmt.Sprintf("| %s | %s | %s |\n",
				formattedTime,
				lesson.SubjectName,
				TypeEducation,
				// lesson.TeacherName,
			)
		}
		message += "\n" // Добавляем пустую строку между днями
	}

	sendMessage(bot, userID, message, authContext)

	//tofoo

}

func getScheduleMessage(requestType, groupName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Выполняем gRPC вызов для получения расписания
	scheduleResp, err := getSchedule(ctx, groupName, requestType)
	if err != nil {
		return fmt.Sprintf("Ошибка при получении расписания. Пожалуйста, попробуйте снова. %s", err)
	}

	// Формируем сообщение с расписанием
	if len(scheduleResp.Lessons) == 0 {
		return "Расписание пусто."
	}

	// Создаем мапу для хранения расписания по дням
	type DaySchedule struct {
		Date    string
		Weekday string
		Lessons []*pbs.Lesson
	}

	schedules := make(map[string]*DaySchedule)

	for _, lesson := range scheduleResp.Lessons {
		// Парсим время начала
		startTime, err := time.Parse(time.RFC3339, lesson.StartTime)
		if err != nil {
			return "Ошибка при обработке даты."
		}

		// Получаем день недели и дату
		weekday := daysOfWeek[startTime.Weekday()]
		formattedDate := startTime.Format("02.01.2006") // Форматируем дату

		// Проверяем, существует ли уже запись для этой даты
		if daySchedule, exists := schedules[formattedDate]; exists {
			// Если запись существует, добавляем урок
			daySchedule.Lessons = append(daySchedule.Lessons, lesson)
		} else {
			// Если записи нет, создаем новую
			schedules[formattedDate] = &DaySchedule{
				Date:    formattedDate,
				Weekday: weekday,
				Lessons: []*pbs.Lesson{lesson},
			}
		}
	}

	// Сортируем расписание по дате
	var sortedSchedules []*DaySchedule
	for _, daySchedule := range schedules {
		sortedSchedules = append(sortedSchedules, daySchedule)
	}
	sort.Slice(sortedSchedules, func(i, j int) bool {
		return sortedSchedules[i].Date < sortedSchedules[j].Date
	})

	// Формируем сообщение
	var message string
	for _, daySchedule := range sortedSchedules {
		message += fmt.Sprintf("🔴 Расписание на %s (%s):\n", daySchedule.Date, daySchedule.Weekday)
		for _, lesson := range daySchedule.Lessons {
			// Парсим время начала и конца
			startTime, _ := time.Parse(time.RFC3339, lesson.StartTime)
			endTime, _ := time.Parse(time.RFC3339, lesson.EndTime)

			// Форматируем время
			formattedTime := fmt.Sprintf("%s - %s", startTime.Format("15:04"), endTime.Format("15:04"))

			// Проверяем SubjectName и заменяем его при необходимости
			TypeEducation := lesson.TypeEducation
			if TypeEducation == "Практическое занятие" {
				TypeEducation = "Практика"
			}
			if TypeEducation == "Лабораторное занятие" {
				TypeEducation = "Лабораторная"
			}

			message += fmt.Sprintf("| %s | %s | %s |\n",
				formattedTime,
				lesson.SubjectName,
				TypeEducation,
				// lesson.TeacherName,
			)
		}
		message += "\n" // Добавляем пустую строку между днями
	}

	return message

}

func sendTeachersInfoMenu(bot *telego.Bot, userID int64, authContext *AuthContext, chatID int64) {

	clearMessages(bot, authContext, chatID)

	if authContext.LastMessageID != 0 {
		err := bot.DeleteMessage(&telego.DeleteMessageParams{
			ChatID:    tu.ID(chatID),
			MessageID: int(authContext.LastMessageID),
		})
		if err != nil {
			log.Printf("Ошибка при удалении сообщения: %v", err)
		}
	}

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Поиск по ФИО преподавателя").WithCallbackData("teacher_find_fio"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Поиск по названию кафедры").WithCallbackData("teacher_find_department"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Поиск по названию предмета").WithCallbackData("teacher_find_subject"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Назад").WithCallbackData("teacher_find_back"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		"Какой критерий вы предпочитаете использовать для поиска преподавателя?",
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		// authContext.LastMessageID = int64(sentMessage.MessageID)
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}

}

func sendDocumentsInfo(bot *telego.Bot, userID int64, authContext *AuthContext) {
	// Логика для отправки информации о шаблонах/бланках документов
	authContext.State = StateDocumentMenu
	sendDocumentsMenu(bot, userID, authContext, userID)
}

func sendExtracurricularInfo(bot *telego.Bot, userID int64, authContext *AuthContext) {
	// Логика для отправки информации о внеурочной активной деятельности
	message := "Здесь будет информация о внеурочной активной деятельности."
	sendMessage(bot, userID, message, authContext)
}

func sendQuestionForm(bot *telego.Bot, userID int64, authContext *AuthContext) {
	// Логика для отправки формы для задания вопроса
	message := "Здесь будет форма для задания вопроса."
	sendMessage(bot, userID, message, authContext)
}

func getGroupAndSendSchedule(bot *telego.Bot, userID int64, requestType string, tgID int64, authContext *AuthContext, chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Получаем название группы по tg_id
	groupName, err := getGroupByTGID(ctx, tgID)
	if err != nil {
		sendMessage(bot, userID, "Ошибка при получении группы. Пожалуйста, попробуйте снова.", authContext)
		return
	}

	// Теперь получаем расписание по названию группы
	message := getScheduleMessage(requestType, groupName)
	toMainMenu2(bot, userID, message, authContext, chatID)
}

func getGroupByTGID(ctx context.Context, tgID int64) (string, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbs.NewScheduleServiceClient(conn)
	resp, err := client.GetGroupByTGID(ctx, &pbs.GetGroupByTGIDRequest{ProfileTgId: tgID})
	if err != nil {
		return "", fmt.Errorf("ошибка при вызове gRPC метода GetGroupByTGID: %v", err)
	}

	return resp.GroupName, nil
}

func getSchedule(ctx context.Context, groupName, requestType string) (*pbs.ScheduleResponse, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbs.NewScheduleServiceClient(conn)
	resp, err := client.GetSchedule(ctx, &pbs.ScheduleRequest{
		GroupName:   groupName,
		RequestType: requestType,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода GetSchedule: %v", err)
	}

	return resp, nil
}

func checkUser(ctx context.Context, userID int64) (*pba.CheckUserResponse, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pba.NewAuthClient(conn)
	resp, err := client.CheckUser(ctx, &pba.CheckUserRequest{ProfileTgId: userID})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода CheckUser: %v", err)
	}

	return resp, nil
}

func registerUser(ctx context.Context, userID int64, profileName, groupName string) (*pba.RegisterUserResponse, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pba.NewAuthClient(conn)
	resp, err := client.RegisterUser(ctx, &pba.RegisterUserRequest{
		ProfileTgId: userID,
		ProfileName: profileName,
		GroupName:   groupName, // Передаем группу
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода RegisterUser: %v", err)
	}

	return resp, nil
}

func getGroups(ctx context.Context) (*pba.GetGroupsResponse, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pba.NewAuthClient(conn)
	resp, err := client.GetGroups(ctx, &pba.Empty{})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода GetGroups: %v", err)
	}

	return resp, nil
}

// Функция для поиска преподавателей по ФИО
func findTeachersByFIO(ctx context.Context, fio string) ([]*pbt.Teacher, error) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbt.NewTeacherServiceClient(conn)
	resp, err := client.FindTeachersByFIO(ctx, &pbt.FindTeachersRequest{Fio: fio})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода FindTeachersByFIO: %v", err)
	}

	return resp.Teachers, nil
}

// Функция для поиска преподавателей по кафедре
func findTeachersByDepartment(ctx context.Context, department string) ([]*pbt.Teacher, error) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbt.NewTeacherServiceClient(conn)
	resp, err := client.FindTeachersByDepartment(ctx, &pbt.FindTeachersByDepartmentRequest{Department: department})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода FindTeachersByDepartment: %v", err)
	}

	return resp.Teachers, nil
}

// Функция для поиска преподавателей по предмету
func findTeachersBySubject(ctx context.Context, subject string) ([]*pbt.Teacher, error) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbt.NewTeacherServiceClient(conn)
	resp, err := client.FindTeachersBySubject(ctx, &pbt.FindTeachersBySubjectRequest{Subject: subject})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода FindTeachersBySubject: %v", err)
	}

	return resp.Teachers, nil
}

func findAddressByPlaceName(ctx context.Context, placeName string) ([]*pbac.Place, error) {
	conn, err := grpc.Dial("localhost:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbac.NewAddressServiceClient(conn)
	resp, err := client.GetAddressInfo(ctx, &pbac.AddressRequest{PlaceName: placeName})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода GetAddressInfo: %v", err)
	}

	return resp.Places, nil
}

func handleAddressButtonPress(bot *telego.Bot, userID int64, placeName string, authContext *AuthContext) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Получаем сообщения по placeName
	messages, err := getMessagesByPlaceName(ctx, placeName)
	if err != nil {
		// Обработка ошибки, если не удалось получить сообщения
		sendMessage(bot, userID, fmt.Sprintf("Ошибка при получении информации об адресе 1: %s", err), authContext)
		return
	}

	authContext.LastPlaceName = placeName

	clearMessages(bot, authContext, userID)

	for _, message := range messages {
		sendMessagesAdress(bot, userID, authContext, message)
	}

}

func getMessagesByPlaceName(ctx context.Context, placeName string) ([]string, error) {
	// Выполняем gRPC вызов для получения информации об адресе
	places, err := findAddressByPlaceName(ctx, placeName)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении информации об адресе: %w", err)
	}

	// Формируем сообщения
	var messages []string
	for _, place := range places {
		// Преобразование времени
		workTime, err := formatWorkTime(place.PlaceTimeStart, place.PlaceTimeEnd)
		if err != nil {
			workTime = "неизвестно" // Обработка ошибки
		}

		// Формирование текста сообщения
		messageText := fmt.Sprintf("Название: %s\nВремя работы: %s\nТелефон: %s\nEmail: %s\nАдрес: %s",
			place.PlaceName, workTime, place.PlacePhone, place.PlaceEmail, place.PlaceAdress)

		messages = append(messages, messageText)
	}

	return messages, nil
}

func sendMessagesAdress(bot *telego.Bot, userID int64, authContext *AuthContext, messageIn string) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Узнать расположение").WithCallbackData("send_location"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		messageIn,
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке меню: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendMessageAdressLocation(bot *telego.Bot, userID int64, authContext *AuthContext, latitude, longitude float64) {
	clearMessages(bot, authContext, userID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
		),
	)

	location := tu.Location(telego.ChatID{
		ID: userID,
	}, latitude, longitude).WithReplyMarkup(inlineKeyboard)

	sentLocation, err := bot.SendLocation(location)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentLocation.MessageID))
	}
}

// Убедитесь, что функция formatWorkTime определена и используется
func formatWorkTime(start, end string) (string, error) {
	// Парсинг времени
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return "", err
	}
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return "", err
	}

	// Форматирование времени в нужный формат
	return fmt.Sprintf("%02d:%02d - %02d:%02d", startTime.Hour(), startTime.Minute(), endTime.Hour(), endTime.Minute()), nil
}

func sendMessage(bot *telego.Bot, userID int64, message string, authContext *AuthContext) {
	clearMessages(bot, authContext, userID)

	sentMessage, err := bot.SendMessage(
		tu.Message(
			tu.ID(userID),
			message,
		),
	)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func sendMessageWithoutDelete(bot *telego.Bot, userID int64, message string, authContext *AuthContext) {

	sentMessage, err := bot.SendMessage(
		tu.Message(
			tu.ID(userID),
			message,
		),
	)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func clearMessages(bot *telego.Bot, authContext *AuthContext, chatID int64) {
	if len(authContext.LastMessageIDs) > 0 {
		for _, messageID := range authContext.LastMessageIDs {
			err := bot.DeleteMessage(&telego.DeleteMessageParams{
				ChatID:    tu.ID(chatID),
				MessageID: int(messageID),
			})
			if err != nil {
				log.Printf("Ошибка при удалении сообщения: %v", err)
			}
		}
		// Очищаем слайс после удаления
		authContext.LastMessageIDs = []int64{}
	}
}
