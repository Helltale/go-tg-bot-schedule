package handlecommand

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"tgclient/internal/clients"
	"tgclient/internal/fileutils"
	"tgclient/internal/messages"
	"tgclient/internal/messagetext"
	"tgclient/internal/models"
	"tgclient/internal/utils"
	pbac "tgclient/proto/adress-contact"
	pbs "tgclient/proto/schedule"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

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

	authContext := &models.AuthContext{State: models.StateStart}

	//start
	bh.Handle(func(b *telego.Bot, update telego.Update) {
		authContext.UserID = update.Message.Chat.ID
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// Проверяем, зарегистрирован ли пользователь
		resp, err := clients.CheckUser(ctx, authContext.UserID)
		if err != nil {
			messages.SendMessage(b, update.Message.Chat.ID, "Ошибка при проверке пользователя. Пожалуйста, попробуйте снова.", authContext)
			return
		}

		if resp.Exists {
			authContext.ProfileName = resp.ProfileName
			authContext.State = models.StateAuthorized
			messages.SendMainMenu(b, update.Message.Chat.ID, authContext)
		} else {
			authContext.State = models.StateAwaitingName
			messages.SendMessageWithoutDelete(b, update.Message.Chat.ID, "Добро пожаловать!\nПожалуйста, введите ваше ФИО:", authContext)
		}
	}, th.CommandEqual("start"))

	bh.Handle(func(b *telego.Bot, update telego.Update) {
		if update.Message != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			switch authContext.State {
			case models.StateStart:
				authContext.UserID = update.Message.Chat.ID
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				resp, err := clients.CheckUser(ctx, authContext.UserID)
				if err != nil {
					messages.SendMessage(b, update.Message.Chat.ID, "Ошибка при проверке пользователя. Пожалуйста, попробуйте снова.", authContext)
					return
				}

				if resp.Exists {
					authContext.ProfileName = resp.ProfileName
					authContext.State = models.StateAuthorized
					messages.SendMessage(b, update.Message.Chat.ID, "Вы успешно авторизованы!", authContext)

					messages.SendScheduleMenu(b, update.Message.Chat.ID, authContext)
				} else {
					messages.SendMessage(b, update.Message.Chat.ID, "Пользователь не найден. Пожалуйста, введите ваше имя:", authContext)
					authContext.State = models.StateAwaitingName
				}
			case models.StateAwaitingName:
				authContext.ProfileName = update.Message.Text
				messages.SendGroupSelection(b, update.Message.Chat.ID, authContext)
				authContext.State = models.StateAwaitingGroup

			case models.StateFindTeacherAwaitingFIO:
				teachers, err := clients.FindTeachersByFIO(ctx, update.Message.Text)
				if err != nil {
					messages.SendMessage(b, update.Message.Chat.ID, fmt.Sprintf("Ошибка при поиске преподавателей: %v", err), authContext)
					return
				}
				message, imgs := messagetext.TextForTeacherInfo(teachers)
				messages.SendTeachersInfo(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID, teachers, message, imgs)
				authContext.State = models.StateToMainMenu // Вернуться

			case models.StateFindTeacherAwaitingDepartment:
				teachers, err := clients.FindTeachersByDepartment(ctx, update.Message.Text)
				if err != nil {
					messages.SendMessage(b, update.Message.Chat.ID, fmt.Sprintf("Ошибка при поиске преподавателей: %v", err), authContext)
					return
				}
				message, imgs := messagetext.TextForTeacherInfo(teachers)
				messages.SendTeachersInfo(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID, teachers, message, imgs)
				authContext.State = models.StateToMainMenu // Вернуться к авторизованному состоянию

			case models.StateFindTeacherAwaitingSubject:
				teachers, err := clients.FindTeachersBySubject(ctx, update.Message.Text)
				if err != nil {
					messages.SendMessage(b, update.Message.Chat.ID, fmt.Sprintf("Ошибка при поиске преподавателей: %v", err), authContext)
					return
				}
				message, imgs := messagetext.TextForTeacherInfo(teachers)
				messages.SendTeachersInfo(b, update.Message.Chat.ID, authContext, update.Message.Chat.ID, teachers, message, imgs)
				authContext.State = models.StateToMainMenu // Вернуться к авторизованному состоянию

			}

		} else if update.CallbackQuery != nil {
			if authContext.State == models.StateAwaitingGroup {
				group := update.CallbackQuery.Data

				messages.SendMessage(b, update.CallbackQuery.From.ID, fmt.Sprintf("Вы выбрали группу: %s", group), authContext)

				// Регистрация пользователя
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := clients.RegisterUser(ctx, authContext.UserID, authContext.ProfileName, group) // Передаем группу
				if err != nil {
					messages.SendMessage(b, update.CallbackQuery.From.ID, "Ошибка при регистрации. Пожалуйста, попробуйте снова.", authContext)
					return
				}
				authContext.State = models.StateAuthorized

				messages.SendMessageWithoutDelete(b, update.CallbackQuery.From.ID, "Вы успешно зарегистрированы!", authContext)
				messages.SendMainMenu(b, update.Message.Chat.ID, authContext)
			} else if authContext.State == models.StateAuthorized {
				// Обработка нажатий на кнопки расписания

				if update.CallbackQuery.Data == "schedule_today" {
					sendSchedule(b, update.CallbackQuery.From.ID, "today", authContext.ProfileName, authContext)

				} else if update.CallbackQuery.Data == "schedule_week" {
					sendSchedule(b, update.CallbackQuery.From.ID, "week", authContext.ProfileName, authContext)

				}
			} else if authContext.State == models.StateToMainMenu {
				// Обработка нажатий на кнопки расписания

				if update.CallbackQuery.Data == "main_menu" {
					messages.SendMainMenu(b, update.Message.Chat.ID, authContext)
				}
			}
		}
	}, th.AnyMessage())

	bh.Handle(func(b *telego.Bot, update telego.Update) {
		if update.CallbackQuery != nil {

			if authContext.State == models.StateAwaitingGroup {
				group := update.CallbackQuery.Data
				messages.SendMessageWithoutDelete(b, update.CallbackQuery.From.ID, fmt.Sprintf("Вы выбрали группу: %s", group), authContext)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err := clients.RegisterUser(ctx, authContext.UserID, authContext.ProfileName, group) // Передаем группу
				if err != nil {
					messages.SendMessageWithoutDelete(b, update.CallbackQuery.From.ID, "Ошибка при регистрации. Пожалуйста, попробуйте снова.", authContext)
					return
				}

				authContext.State = models.StateAuthorized
				messages.SendMessageWithoutDelete(b, update.CallbackQuery.From.ID, "Вы успешно зарегистрированы и авторизованы!", authContext)
				message := fmt.Sprintf(`Чтобы быть в курсе последних новостей, скорее перейдите по ссылке, если ещё не подписаны: https://t.me/pmishSamSMU 
	Буду рад ответить на любые ваши вопросы, %s!`, authContext.ProfileName)

				messages.SendMessageWithoutDelete(b, update.CallbackQuery.From.ID, message, authContext)
				messages.SendMainMenuWithoutDelete(b, update.CallbackQuery.From.ID, authContext)

			}

			if authContext.State == models.StateAuthorized {

				switch update.CallbackQuery.Data {
				case "schedule":

					authContext.State = models.StateScheduleMenu
					messages.SendScheduleMenu(b, update.CallbackQuery.From.ID, authContext)
				case "teachers":
					authContext.State = models.StateFindTeacherMenu
					messages.SendTeachersInfoMenu(b, update.CallbackQuery.From.ID, authContext)
				case "contacts":
					authContext.State = models.StateAdressContactMenu
					messages.SendAdressContactMenu(bot, update.CallbackQuery.From.ID, authContext)
				case "documents":
					authContext.State = models.StateDocumentMenu
					messages.SendDocumentsMenu(bot, update.CallbackQuery.From.ID, authContext)
					// case "extracurricular":
					// 	// Обработка нажатия на "Внеурочная активная деятельность"
					// 	sendExtracurricularInfo(b, update.CallbackQuery.From.ID, authContext)
					// case "ask_question":
					// 	// Обработка нажатия на "Задать вопрос"
					// 	sendQuestionForm(b, update.CallbackQuery.From.ID, authContext)
					// case "change_user":
					// 	// Обработка нажатия на "Сменить пользователя"
					// 	authContext.State = models.StateStart
					// 	messages.SendMessage(b, update.CallbackQuery.From.ID, "Вы вышли из аккаунта. Пожалуйста, введите ваше имя:", authContext)
				}
			}

			if authContext.State == models.StateScheduleMenu {
				// Обработка нажатий на кнопки расписания
				if update.CallbackQuery.Data == "schedule_today" {
					authContext.State = models.StateToMainMenu
					getGroupAndSendSchedule(b, update.CallbackQuery.From.ID, "today", authContext.UserID, authContext)

				} else if update.CallbackQuery.Data == "schedule_week" {
					authContext.State = models.StateToMainMenu
					getGroupAndSendSchedule(b, update.CallbackQuery.From.ID, "week", authContext.UserID, authContext)

				} else if update.CallbackQuery.Data == "schedule_back" { //кнопка назад (к главному меню) из выбора расписания расписания
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized
				}
			}

			if authContext.State == models.StateToMainMenu {
				if update.CallbackQuery.Data == "main_menu" {
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized
				}
			}

			if authContext.State == models.StateFindTeacherMenu {
				// Обработка нажатий на маню посика препода
				if update.CallbackQuery.Data == "teacher_find_back" {
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized
				} else if update.CallbackQuery.Data == "teacher_find_fio" {
					messages.SendMessage(b, update.CallbackQuery.From.ID, "Введите фио преподавателя:", authContext)
					authContext.State = models.StateFindTeacherAwaitingFIO
				} else if update.CallbackQuery.Data == "teacher_find_department" {
					messages.SendMessage(b, update.CallbackQuery.From.ID, "Введите кафедру преподавателя:", authContext)
					authContext.State = models.StateFindTeacherAwaitingDepartment
				} else if update.CallbackQuery.Data == "teacher_find_subject" {
					messages.SendMessage(b, update.CallbackQuery.From.ID, "Введите предмет:", authContext)
					authContext.State = models.StateFindTeacherAwaitingSubject
				}

			}

			if authContext.State == models.StateAdressContactMenu {
				if update.CallbackQuery.Data == "adress_contact_back" {
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized

				} else if update.CallbackQuery.Data == "adress_contact_administrative" {
					place := "Административный корпус"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = models.StateToMainMenu

				} else if update.CallbackQuery.Data == "adress_contact_study" {
					place := "Учебный корпус"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = models.StateToMainMenu

				} else if update.CallbackQuery.Data == "adress_contact_living" {
					place := "Общежитие"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = models.StateToMainMenu

				} else if update.CallbackQuery.Data == "adress_contact_departments" {
					place := "Кафедры"
					handleAddressButtonPress(b, update.CallbackQuery.From.ID, place, authContext)
					authContext.SelectedPlace = &pbac.Place{PlaceName: place} // Сохраняем выбранное место
					authContext.State = models.StateToMainMenu

				}
			}

			if authContext.State == models.StateToMainMenu {
				if update.CallbackQuery.Data == "send_location" {

					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()

					if authContext.SelectedPlace != nil {
						// Получаем информацию о месте
						places, err := clients.FindAddressByPlaceName(ctx, authContext.SelectedPlace.PlaceName) // Используем сохраненное место
						if err != nil {
							log.Printf("Ошибка при получении адреса: %v", err)
							return
						}

						// Предполагаем, что вы хотите отправить локацию первого места
						if len(places) > 0 {
							latitude := places[0].PlaceAdressPoint.Latitude
							longitude := places[0].PlaceAdressPoint.Longitude
							messages.SendMessageAdressLocation(bot, update.CallbackQuery.From.ID, authContext, latitude, longitude)
						} else {
							log.Println("Нет доступных мест для отправки локации.")
						}
					} else {
						log.Println("Нет выбранного места.")
					}

				}
			}

			if authContext.State == models.StateDocumentMenu {
				if update.CallbackQuery.Data == "documents_back" {
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized
				} else if update.CallbackQuery.Data == "document_group_1" {
					messages.SendDocumentsGroup1Menu(b, update.CallbackQuery.From.ID, authContext, update.CallbackQuery.From.ID)
					authContext.State = models.StateDocumentGroup1Menu
				}
			}

			if authContext.State == models.StateDocumentGroup1Menu {
				if update.CallbackQuery.Data == "document_group_1_doc1" {
					message := fileutils.GetFileInfo("doc1")
					messages.SendFileInfoMessage(b, update.CallbackQuery.From.ID, message, authContext)
					authContext.State = models.StateReadyForDownloadDocument

				} else if update.CallbackQuery.Data == "document_group_1_doc2" {
					messages.SendMessage(b, update.CallbackQuery.From.ID, "doc2", authContext)
				} else if update.CallbackQuery.Data == "document_group_1_doc3" {
					messages.SendMessage(b, update.CallbackQuery.From.ID, "doc3", authContext)
				} else if update.CallbackQuery.Data == "main_menu" {
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized
				}
			}

			if authContext.State == models.StateReadyForDownloadDocument {
				if update.CallbackQuery.Data == "main_menu" {
					messages.SendMainMenu(b, update.CallbackQuery.From.ID, authContext)
					authContext.State = models.StateAuthorized
				} else if update.CallbackQuery.Data == "downloading_file" {
					messages.SendFileInfoDocument(b, update.CallbackQuery.From.ID, authContext)
				}
			}

		}
	}, th.AnyCallbackQuery())

	bh.Start()
}

func sendSchedule(bot *telego.Bot, userID int64, requestType, groupName string, authContext *models.AuthContext) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Выполняем gRPC вызов для получения расписания
	scheduleResp, err := clients.GetSchedule(ctx, groupName, requestType)
	if err != nil {
		messages.SendMessageInlineKeyboard(bot, userID, authContext, fmt.Sprintf("Ошибка при получении расписания. Пожалуйста, попробуйте снова. %s", err), "В главное меню", "main_menu")
		return
	}

	// Формируем сообщение с расписанием
	if len(scheduleResp.Lessons) == 0 {
		messages.SendMessageInlineKeyboard(bot, userID, authContext, "Расписание пусто.", "В главное меню", "main_menu")
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
			messages.SendMessageInlineKeyboard(bot, userID, authContext, "Ошибка при обработке даты.", "В главное меню", "main_menu")
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

	messages.SendMessage(bot, userID, message, authContext)

	//tofoo

}

func getScheduleMessage(requestType, groupName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Выполняем gRPC вызов для получения расписания
	scheduleResp, err := clients.GetSchedule(ctx, groupName, requestType)
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

// func sendExtracurricularInfo(bot *telego.Bot, userID int64, authContext *AuthContext) {
// 	// Логика для отправки информации о внеурочной активной деятельности
// 	message := "Здесь будет информация о внеурочной активной деятельности."
// 	sendMessage(bot, userID, message, authContext)
// }

// func sendQuestionForm(bot *telego.Bot, userID int64, authContext *AuthContext) {
// 	// Логика для отправки формы для задания вопроса
// 	message := "Здесь будет форма для задания вопроса."
// 	sendMessage(bot, userID, message, authContext)
// }

func getGroupAndSendSchedule(bot *telego.Bot, userID int64, requestType string, tgID int64, authContext *models.AuthContext) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Получаем название группы по tg_id
	groupName, err := clients.GetGroupByTGID(ctx, tgID)
	if err != nil {
		messages.SendMessageInlineKeyboard(bot, userID, authContext, "Ошибка при получении группы. Пожалуйста, попробуйте снова.", "В главное меню", "main_menu")
		return
	}

	// Теперь получаем расписание по названию группы
	message := getScheduleMessage(requestType, groupName)
	messages.SendMessageInlineKeyboard(bot, userID, authContext, message, "В главное меню", "main_menu")
}

func handleAddressButtonPress(bot *telego.Bot, userID int64, placeName string, authContext *models.AuthContext) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Получаем сообщения по placeName
	message, err := getMessagesByPlaceName(ctx, placeName)
	if err != nil {
		// Обработка ошибки, если не удалось получить сообщения
		messages.SendMessageInlineKeyboard(bot, userID, authContext, fmt.Sprintf("Ошибка при получении информации об адресе 1: %s", err), "В главное меню", "main_menu")
		return
	}

	authContext.LastPlaceName = placeName

	messages.ClearMessages(bot, authContext, userID)

	for _, messag := range message {
		messages.SendMessageAdress(bot, userID, authContext, messag)
	}

}

func getMessagesByPlaceName(ctx context.Context, placeName string) ([]string, error) {
	// Выполняем gRPC вызов для получения информации об адресе
	places, err := clients.FindAddressByPlaceName(ctx, placeName)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении информации об адресе: %w", err)
	}

	// Формируем сообщения
	var message []string
	for _, place := range places {
		// Преобразование времени
		workTime, err := utils.FormatWorkTime(place.PlaceTimeStart, place.PlaceTimeEnd)
		if err != nil {
			workTime = "неизвестно" // Обработка ошибки
		}

		// Формирование текста сообщения
		messageText := fmt.Sprintf("Название: %s\nВремя работы: %s\nТелефон: %s\nEmail: %s\nАдрес: %s",
			place.PlaceName, workTime, place.PlacePhone, place.PlaceEmail, place.PlaceAdress)

		message = append(message, messageText)
	}

	return message, nil
}
