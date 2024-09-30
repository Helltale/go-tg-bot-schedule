package models

import pbac "tgclient/proto/adress-contact"

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
	StateDocumentGroup2Menu
	StateDocumentGroup3Menu
	StateReadyForDownloadDocument1 //готовность для скачивания файла
	StateReadyForDownloadDocument2
	StateReadyForDownloadDocument3
	StateReadyForDownloadDocument4
	StateReadyForDownloadDocument5
	StateReadyForDownloadDocument6
	StateReadyForDownloadDocument7
	StateReadyForDownloadDocument8
	StateReadyForDownloadDocument9

	StateTeacherMainMenu
	StateTeacherEmailMessage
	StateTeacherPassMessage
	StateTeacherVacationMessage
	StateTeacherVacationSelfMessage
	StateTeacherReferenseMessage
	StateTeacherPaySheetMessage
	StateTeacherMedicalMessage
	StateTeacherHelpDesk
)

type AuthContext struct {
	State          AuthState
	UserID         int64
	ProfileName    string
	ProfileRole    string
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
