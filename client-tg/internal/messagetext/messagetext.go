package messagetext

import (
	"fmt"
	pbt "tgclient/proto/teacher"
)

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
