package telegram

import (
	"github.com/Syfaro/telegram-bot-api"
	"golang.org/x/net/context"
)

func mapRouteHandler(pattern string, handlers []Handler) []Handler {
	var mapped []Handler
	for _, handler := range handlers {
		mapped = append(mapped, &Route{Equal{pattern}, handler})
	}

	return mapped
}

func mapHandlerFunc(handlers []func(context.Context, *Control, *tgbotapi.BotAPI, *tgbotapi.Update)) []Handler {
	var mapped []Handler
	for _, handler := range handlers {
		mapped = append(mapped, HandlerFunc(handler))
	}

	return mapped
}

func mapErrorHandlerFunc(handlers []func(context.Context, *Control, *tgbotapi.BotAPI, *tgbotapi.Update, error)) []ErrorHandler {
	var mapped []ErrorHandler
	for _, handler := range handlers {
		mapped = append(mapped, ErrorHandlerFunc(handler))
	}

	return mapped
}
