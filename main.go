package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"lang.bot/scheduler"
	env "lang.bot/environ"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

)

type TranslateResult struct {
	text	string
	err		error
}

func main() {
	envs := env.GetEnvironments()
	if envs.BotToken == "" {
		panic("TOKEN environment variable is empty")
	}

	b, err := gotgbot.NewBot(envs.BotToken, &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			Client: http.Client{},
			DefaultRequestOpts: &gotgbot.RequestOpts{
				Timeout: gotgbot.DefaultTimeout,
				APIURL:  gotgbot.DefaultAPIURL,
			},
		},
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	c := &client{}

	dispatcher.AddHandler((handlers.NewCommand("start", start)))

	dispatcher.AddHandler(handlers.NewCommand("choose_source_lang", choose_source_lang))
	dispatcher.AddHandler(handlers.NewCommand("choose_target_lang", choose_target_lang))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("source_lang"), c.source_lang_callback))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("target_lang"), c.target_lang_callback))

	dispatcher.AddHandler(handlers.NewMessage(message.Text, c.translate))

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)

	go scheduler.Start_scheduler()

	updater.Idle()
}

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveSender.ChatId, fmt.Sprintf("Hello, I'm @%s.\nI am a bot that can help you translate text to a desired language, see commands for more info", b.User.Username), &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func choose_target_lang(b *gotgbot.Bot, ctx *ext.Context) error {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Russian", CallbackData: "target_lang_ru"},
				{Text: "English", CallbackData: "target_lang_en"},
			},
			{
				{Text: "French", CallbackData: "target_lang_fr"},
				{Text: "German", CallbackData: "target_lang_de"},
			},
			{
				{Text: "Irish", CallbackData: "target_lang_ga"},
				{Text: "Japanese", CallbackData: "target_lang_ja"},
			},
			{
				{Text: "Polish", CallbackData: "target_lang_pl"},
				{Text: "Chinese", CallbackData: "target_lang_zh"},
			},
		},
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, "Please choose a target language:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}

	return err
}

func choose_source_lang(b *gotgbot.Bot, ctx *ext.Context) error {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Russian", CallbackData: "source_lang_ru"},
				{Text: "English", CallbackData: "source_lang_en"},
			},
			{
				{Text: "French", CallbackData: "source_lang_fr"},
				{Text: "German", CallbackData: "source_lang_de"},
			},
			{
				{Text: "Irish", CallbackData: "source_lang_ga"},
				{Text: "Japanese", CallbackData: "source_lang_ja"},
			},
			{
				{Text: "Polish", CallbackData: "source_lang_pl"},
				{Text: "Chinese", CallbackData: "source_lang_zh"},
			},
			{
				{Text: "Auto-detection", CallbackData: "source_lang_auto"},
			},
		},
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, "Please choose a source language:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}

	return err
}