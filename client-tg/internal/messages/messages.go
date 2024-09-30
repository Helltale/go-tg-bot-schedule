package messages

import (
	"context"
	"fmt"
	"log"
	"os"
	"tgclient/internal/clients"
	"tgclient/internal/fileutils"
	"tgclient/internal/models"
	"time"

	pbt "tgclient/proto/teacher"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func SendMessageWithoutDelete(bot *telego.Bot, userID int64, message string, authContext *models.AuthContext) {

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

func SendMessage(bot *telego.Bot, userID int64, message string, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

	SendMessageWithoutDelete(bot, userID, message, authContext)
}

func SendMessageInlineKeyboardWithoutDelete(bot *telego.Bot, userID int64, authContext *models.AuthContext, messageText string, buttonText string, buttonCallbackData string) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton(buttonText).WithCallbackData(buttonCallbackData),
		),
	)

	message := tu.Message(
		tu.ID(userID),
		messageText,
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения (id: `%d` 	text: `%s`) : %v", sentMessage.MessageID, sentMessage.Text, err)
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func SendMessageInlineKeyboard(bot *telego.Bot, userID int64, authContext *models.AuthContext, messageText string, buttonText string, buttonCallbackData string) {

	ClearMessages(bot, authContext, userID)

	SendMessageInlineKeyboardWithoutDelete(bot, userID, authContext, messageText, buttonText, buttonCallbackData)
}

func SendMainMenuWithoutDelete(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Преподаватели").WithCallbackData("teachers"),
			tu.InlineKeyboardButton("Расписание").WithCallbackData("schedule"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Адреса и контакты").WithCallbackData("contacts"),
			tu.InlineKeyboardButton("Бланки документов").WithCallbackData("documents"),
		),
		// tu.InlineKeyboardRow(я
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

func SendMainMenu(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

	SendMainMenuWithoutDelete(bot, userID, authContext)
}

func SendMainMenuWithoutDeleteHybrid(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Преподаватели").WithCallbackData("teachers"),
			tu.InlineKeyboardButton("Расписание").WithCallbackData("schedule"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Адреса и контакты").WithCallbackData("contacts"),
			tu.InlineKeyboardButton("Бланки документов").WithCallbackData("documents"),
		),
		// tu.InlineKeyboardRow(я
		// 	tu.InlineKeyboardButton("Внеурочная активная деятельность").WithCallbackData("extracurricular"),
		// 	tu.InlineKeyboardButton("Задать вопрос").WithCallbackData("ask_question"),
		// ),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Сменить роль").WithCallbackData("hybrid_to_teacher"),
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
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func SendMainMenuHybrid(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

	SendMainMenuWithoutDeleteHybrid(bot, userID, authContext)
}

func SendScheduleMenuWithoutDelete(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

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
	}
}

func SendScheduleMenu(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

	SendScheduleMenuWithoutDelete(bot, userID, authContext)
}

func ClearMessages(bot *telego.Bot, authContext *models.AuthContext, chatID int64) {
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

		authContext.LastMessageIDs = []int64{}
	}
}

func SendMessageAdress(bot *telego.Bot, userID int64, authContext *models.AuthContext, messageIn string) {

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

func SendMessageAdressLocation(bot *telego.Bot, userID int64, authContext *models.AuthContext, latitude, longitude float64) {
	ClearMessages(bot, authContext, userID)

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

func SendDocumentsMenu(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

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

func SendDocumentsGroup1Menu(bot *telego.Bot, userID int64, authContext *models.AuthContext, chatID int64) {

	ClearMessages(bot, authContext, chatID)

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

func SendDocumentsGroup2Menu(bot *telego.Bot, userID int64, authContext *models.AuthContext, chatID int64) {

	ClearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 4").WithCallbackData("document_group_2_doc4"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 5").WithCallbackData("document_group_2_doc5"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 6").WithCallbackData("document_group_2_doc6"),
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

func SendDocumentsGroup3Menu(bot *telego.Bot, userID int64, authContext *models.AuthContext, chatID int64) {

	ClearMessages(bot, authContext, chatID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 7").WithCallbackData("document_group_3_doc7"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 8").WithCallbackData("document_group_3_doc8"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Документ 9").WithCallbackData("document_group_3_doc9"),
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

func SendAdressContactMenu(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

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
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func SendFileInfoMessage(bot *telego.Bot, userID int64, messageIn string, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

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

func SendFileInfoDocument(bot *telego.Bot, userID int64, authContext *models.AuthContext, file string, filetype string) {

	ClearMessages(bot, authContext, userID)

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("В главное меню").WithCallbackData("main_menu"),
		),
	)

	var file_ *os.File

	switch filetype {
	case "pdf":
		file_ = fileutils.MustOpenPDF(file)
	case "docx":
		file_ = fileutils.MustOpenDOCX(file)
	default:
		log.Printf("некорретно заданный файл или тип файла для загрузки %s - %s", file, filetype)
		SendMessageInlineKeyboard(bot, userID, authContext, fmt.Sprintf("некорретно заданный файл для загрузки %s", filetype), "В главное меню", "main_menu")
	}

	document := tu.Document(
		tu.ID(userID),
		tu.File(file_),
	).WithReplyMarkup(inlineKeyboard)

	sentMessage, err := bot.SendDocument(document)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func SendTeachersInfo(bot *telego.Bot, userID int64, authContext *models.AuthContext, chatID int64, teachers []*pbt.Teacher, messages []string, imgs []string) {

	ClearMessages(bot, authContext, chatID)

	if len(messages) == 0 {
		SendMessageInlineKeyboard(bot, userID, authContext, "Никого не найдено :с", "В главное меню", "main_menu")
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
			tu.File(fileutils.MustOpenJPG(imgs[index])),
		).WithReplyMarkup(inlineKeyboard).WithCaption(message)

		sentMessage, err := bot.SendPhoto(photo)
		if err != nil {
			log.Printf("Ошибка при отправке результата поиска: %v", err)
		} else {
			authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
		}
	}

}

func SendTeachersInfoMenu(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

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
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

// todo сделать как тут динамическое добавление кнопок
func SendGroupSelection(bot *telego.Bot, userID int64, authContext *models.AuthContext) {
	ClearMessages(bot, authContext, userID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	groupsResp, err := clients.GetGroups(ctx)
	if err != nil {
		SendMessageInlineKeyboard(bot, userID, authContext, "Ошибка при получении групп. Пожалуйста, попробуйте снова.", "В главное меню", "main_menu")
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

func SendTeacherMainMenuWithoutDelete(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Рабочая почта").WithCallbackData("teacher_menu_email"),
			tu.InlineKeyboardButton("Пропуск").WithCallbackData("teacher_menu_pass"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Отпуск").WithCallbackData("teacher_menu_vacation"),
			tu.InlineKeyboardButton("Отпуск за свой счет").WithCallbackData("teacher_menu_vacation_self"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Справка с места работы").WithCallbackData("teacher_menu_reference"),
			tu.InlineKeyboardButton("Расчетный лист").WithCallbackData("teacher_menu_pay_sheet"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Больничный").WithCallbackData("teacher_menu_medical"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Help desk").WithCallbackData("teacher_menu_help_desk"),
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
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func SendTeacherMainMenu(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

	SendTeacherMainMenuWithoutDelete(bot, userID, authContext)
}

func SendTeacherMainMenuWithoutDeleteHybrid(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Рабочая почта").WithCallbackData("teacher_menu_email"),
			tu.InlineKeyboardButton("Пропуск").WithCallbackData("teacher_menu_pass"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Отпуск").WithCallbackData("teacher_menu_vacation"),
			tu.InlineKeyboardButton("Отпуск за свой счет").WithCallbackData("teacher_menu_vacation_self"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Справка с места работы").WithCallbackData("teacher_menu_reference"),
			tu.InlineKeyboardButton("Расчетный лист").WithCallbackData("teacher_menu_pay_sheet"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Больничный").WithCallbackData("teacher_menu_medical"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Help desk").WithCallbackData("teacher_menu_help_desk"),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Сменить роль").WithCallbackData("hybrid_to_student"),
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
		authContext.LastMessageIDs = append(authContext.LastMessageIDs, int64(sentMessage.MessageID))
	}
}

func SendTeacherMainMenuHybrid(bot *telego.Bot, userID int64, authContext *models.AuthContext) {

	ClearMessages(bot, authContext, userID)

	SendTeacherMainMenuWithoutDeleteHybrid(bot, userID, authContext)
}
