package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RabbitMQClient struct {
}

func (r *RabbitMQClient) GetNumOfMessages(url, credentials string) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return -1, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", credentials))

	res, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	var result map[string]interface{}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return -1, err
	}

	return int(result["messages"].(float64)), nil
}
