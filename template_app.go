package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Системные параметры, которые я фильтрую.
// Нужны чтобы исключить их body нового запроса
// так как в моем инстансе количество кастомных и локальных полей перевалило за 500, решил пойти логичным и простым путем
// то есть, от обратного
var systemParameters = map[string]bool{
	"self": true, "id": true, "key": true, "version": true,
	"statusStartTime": true, "statusType": true, "previousStatusLastAssignee": true, "createdAt": true,
	"commentWithExternalMessageCount": true, "updatedAt": true, "lastCommentUpdatedAt": true, "updatedBy": true,
	"sla": true, "followers": true, "createdBy": true, "commentWithoutExternalMessageCount": true,
	"unique": true, "votes": true, "assignee": true, "status": true,
	"previousStatus": true, "favorite": true, "lastQueue": true, "previousQueue": true, "aliases": true, "data": true,
}

const URL string = "https://api.tracker.yandex.net/v3/issues/"

// Для корректно работы необходимо
// именить установленные переменные окружения 
// ORG_ID и TOKEN

func main() {
	// Опционально, заменить на чтение из queryParams
	var issueKey string = "KEY-1"

	// Создаем клиент для работы, позже вынесу в отдельную папку
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	// Запрашиваем информацию о задаче, которую планируем копировать
	request, err := http.NewRequest("GET", fmt.Sprintf(URL+issueKey), nil)
	if err != nil {
		log.Print(err)
	}

	// Добавляем заголовки
	request.Header.Add("Host", "api.tracker.yandex.net")
	request.Header.Add("Authorization", "OAuth "+os.Getenv("TOKEN"))
	request.Header.Add("X-Org-ID", os.Getenv("ORG_ID")) // X-Cloud-Org-ID, если у вас облачный трекер
	request.Header.Add("Content-Type", "application/json")

	// Выполняем запрос, записываем ответ
	response, err := client.Do(request)
	if err != nil {
		log.Print(err)
	}

	// база
	defer response.Body.Close()

	// Превращаем в массив байт для дальнейшней десериализацией
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Print(err)
	}

	// создаем мапу для того, чтобы хранить в ней параметры исходной задачи до фильтрации
	parameters := make(map[string]interface{})

	// десериализуем
	err = json.Unmarshal(body, &parameters)
	if err != nil {
		log.Print(err)
	}

	// map для нового запроса, по сути - тело с параметрами
	filteredParams := make(map[string]interface{})

	// основная логика
	for param, value := range parameters {
		if systemParameters[param] {
			continue
		} else {
			filteredParams[param] = value
		}
	}

	//  Можно было бы записа так, но я решил, что вариант выше читается лучше
	//	for param, value := range parameters {
	//	if !systemParameters[param] {
	//			filteredParams[param] = value
	//		}
	//  }

	// Превращаем в JSON
	encodedBody, err := json.Marshal(filteredParams)
	if err != nil {
		log.Print(err)
	}

	// Формируем новый запрос
	newRequest, err := http.NewRequest("POST", URL, bytes.NewBuffer(encodedBody))
	if err != nil {
		log.Print(err)
	}

	// Тут понятно, пишем заголовки
	newRequest.Header.Add("Authorization", "OAuth "+os.Getenv("TOKEN"))
	newRequest.Header.Add("X-Org-ID", os.Getenv("ORG_ID"))
	newRequest.Header.Add("Content-Type", "application/json")
  
	// Записываем ответ
	response, err = client.Do(newRequest)
	if err != nil || response.StatusCode != 200 {
		log.Printf("Error: %v // Code: %v", err, response.StatusCode)
	}

	defer response.Body.Close()

	log.Print("finished")
}
