package main

import (
	"fideliy/dins"
	"fideliy/helpers"
	"fmt"
	telegram "github.com/acteek/telegram-bot-api"
	"log"
	"strings"
)

const (
	//botToken     = "880777116:AAEOXE6RVHEWzzC0hthVPjwG37WxsUbRy2U" // Prod Bot
	botToken     = "987354230:AAGoDDLMxwowUY_wbuz6UCdgtD33eQE_Q4o" // Test_Bot
	tgEndpoint   = "http://157.230.184.220/bot%s/%s"               //proxy to api.telegram.com
	dinsEndpoint = "https://my.dins.ru"
)

func main() {
	log.Println("Starting...")

	users := make(map[int64]dins.User)
	baskets := make(map[int64][]string)

	bot, err := telegram.NewBotAPI(botToken, tgEndpoint)
	dinsApi := dins.NewDinsApi(dinsEndpoint)
	if err != nil {
		log.Panic("Failed connect to telegram", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	//TODO only for tests
	bot.Debug = true
	users[149199925] = dins.User{ID: "1092", Name: "Sergey Ryazanov", Token: "6ae11e1d81b202ead1733354dce71ba7"}

	u := telegram.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	// TODO   migrate logic to Handler
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				msg := telegram.NewMessage(update.Message.Chat.ID, "")
				switch update.Message.Command() {
				case "set_token":
					m := strings.Split(update.Message.Text, " ")
					if len(m) == 2 {
						token := m[1]
						user, er := dinsApi.GetUser(token)
						if er != nil {
							msg.Text = "Что то пошло не так попробуй другой"
						} else {
							users[update.Message.Chat.ID] = user
							msg.Text = user.Name + ", добро пожаловать"
						}

					} else {
						msg.Text = "Используй команду: /set_token your-token"
					}
				case "start":
					msg.Text = "Дороу !"
					msg.ReplyMarkup = helpers.BuildMainKeyboard()
					fmt.Println(update.Message.Chat.ID)
				default:
					msg.Text = "Я не знаю такой команды"
				}
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

			} else {
				msg := telegram.NewMessage(update.Message.Chat.ID, "")
				switch update.Message.Text {
				case "Меню":
					if user, isAuth := users[update.Message.Chat.ID]; isAuth {
						menu := dinsApi.GetMenu(user)
						if len(menu) == 0 {
							msg.Text = "Сейчас меню не доступно, попробуй позже"
						} else {
							msg.Text = "Вооот"
							msg.ReplyMarkup = helpers.BuildMenuKeyBoard(menu)
						}

					} else {
						msg.Text = "Ты кто такой ...? Используй: /set_token your-token"
					}

				case "Мои заказы":
					msg.Text = "Это пока не реализовано"
					msg.ReplyMarkup = telegram.NewInlineKeyboardMarkup(
						telegram.NewInlineKeyboardRow(
							telegram.NewInlineKeyboardButtonURL("Проверить на Сайте", dinsEndpoint+"/?page=fidel")))
				default:
					msg.Text = "🙀😴"
				}

				if _, err := bot.Send(msg); err != nil {
					log.Panic("Failed Send message", err)
				}
			}
		} else if update.CallbackQuery != nil {

			switch update.CallbackQuery.Data {
			case "make_order":
				orderVieW := strings.Join(baskets[update.CallbackQuery.Message.Chat.ID], ", ")
				msg := telegram.NewMessage(update.CallbackQuery.Message.Chat.ID, orderVieW)
				msg.ReplyMarkup = helpers.BuildOrderKeyBoard()

				sss := telegram.NewDeleteMessage(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID)

				if _, err := bot.Send(msg); err != nil {
					log.Panic("Failed Send message", err)
				}
				if _, err := bot.Send(sss); err != nil {
					log.Panic("Failed Send message", err)
				}
			case "send_order":
				msg := telegram.NewMessage(update.CallbackQuery.Message.Chat.ID, "Заказал для тебя")
				sss := telegram.NewDeleteMessage(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID)

				delete(baskets, update.CallbackQuery.Message.Chat.ID)

				if _, err := bot.Send(msg); err != nil {
					log.Panic("Failed Send message", err)
				}
				if _, err := bot.Send(sss); err != nil {
					log.Panic("Failed Send message", err)
				}
			case "clear_order":
				msg := telegram.NewMessage(update.CallbackQuery.Message.Chat.ID, "Штош ...")
				sss := telegram.NewDeleteMessage(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID)

				delete(baskets, update.CallbackQuery.Message.Chat.ID)

				if _, err := bot.Send(msg); err != nil {
					log.Panic("Failed Send message", err)
				}
				if _, err := bot.Send(sss); err != nil {
					log.Panic("Failed Send message", err)
				}

			default:
				baskets[update.CallbackQuery.Message.Chat.ID] =
					append(baskets[update.CallbackQuery.Message.Chat.ID], update.CallbackQuery.Data)

				if _, err := bot.AnswerCallbackQuery(telegram.NewCallbackWithAlert(update.CallbackQuery.ID, "Добавил в корзину")); err != nil {
					log.Panic("Failed Send message", err)
				}
			}

		}

	}

}
