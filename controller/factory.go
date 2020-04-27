package controller

import "errors"

func Factory(controllerType, conf string) (*Controller, error) {
	switch controllerType {
	case "sqs":
		return NewSQSController(conf)
	case "pubsub":
		return NewPubSubController(conf)
	default:
		return nil, errors.New("invalid controller type")
	}
}
