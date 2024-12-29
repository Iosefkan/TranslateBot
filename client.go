package main

import (
	"sync"
	"strings"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	htgotts "github.com/hegedustibor/htgo-tts"
    handlers "github.com/hegedustibor/htgo-tts/handlers"
	gt "github.com/bas24/googletranslatefree"
)

// A basic handler client to share state across executions.
// Note: This is a very simple layout which uses a shared mutex.
// It is all in-memory, and so will not persist data across restarts.
type client struct {
	// Use a mutex to avoid concurrency issues.
	// If you use multiple maps, you may want to use a new mutex for each one.
	rwMux sync.RWMutex

	// We use a double map to:
	// - map once for the user id
	// - map a second time for the keys a user can have
	// The second map has values of type "any" so anything can be stored in them, for the purpose of this example.
	// This could be improved by using a struct with typed fields, though this would need some additional handling to
	// ensure concurrent safety.
	userData map[int64]map[string]string

	// This struct could also contain:
	// - pointers to database connections
	// - pointers cache connections
	// - localised strings
	// - helper methods for retrieving/caching chat settings
}

func (c *client) getUserData(ctx *ext.Context, key string) (string, bool) {
	c.rwMux.RLock()
	defer c.rwMux.RUnlock()

	if c.userData == nil {
		return "", false
	}

	userData, ok := c.userData[ctx.EffectiveUser.Id]
	if !ok {
		return "", false
	}

	v, ok := userData[key]
	return v, ok
}

func (c *client) setUserData(ctx *ext.Context, key string, val string) {
	c.rwMux.Lock()
	defer c.rwMux.Unlock()

	if c.userData == nil {
		c.userData = map[int64]map[string]string{}
	}

	_, ok := c.userData[ctx.EffectiveUser.Id]
	if !ok {
		c.userData[ctx.EffectiveUser.Id] = map[string]string{}
	}
	c.userData[ctx.EffectiveUser.Id][key] = val
}

func (c *client) source_lang_callback(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.Update.CallbackQuery.Answer(b, nil)
	if err != nil {
		log.Printf("Failed to answer callback query: %v", err)
		return err
	}

	if (strings.Contains(ctx.CallbackQuery.Data, "auto")){
		c.setUserData(ctx, "source_lang", "auto")
	} else {
		lang := ctx.CallbackQuery.Data[len(ctx.CallbackQuery.Data) - 2:]
		c.setUserData(ctx, "source_lang", lang)
	}

	b.SendMessage(ctx.EffectiveSender.ChatId, "Source language updated.", &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})

	return err
}

func (c *client) target_lang_callback(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.Update.CallbackQuery.Answer(b, nil)
	if err != nil {
		log.Printf("Failed to answer callback query: %v", err)
		return err
	}

	lang := ctx.CallbackQuery.Data[len(ctx.CallbackQuery.Data) - 2:]
	c.setUserData(ctx, "target_lang", lang)

	b.SendMessage(ctx.EffectiveSender.ChatId, "Target language updated.", &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})

	return err
}

func (c *client) translate(b *gotgbot.Bot, ctx *ext.Context) error {
	translate_chan := make(chan TranslateResult)

	source_lang, ok := c.getUserData(ctx, "source_lang")
	if !ok {
		source_lang = "auto"
	}

	target_lang, ok := c.getUserData(ctx, "target_lang")
	if !ok {
		target_lang = "en"
	}
	
	go translate_message(ctx.EffectiveMessage.Text, source_lang, target_lang, translate_chan)
	result := <- translate_chan

	if result.err != nil {
		return fmt.Errorf("failed to translate message: %w", result.err)
	}


	_, err := ctx.EffectiveMessage.Reply(b, result.text, nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if target_lang == "ga" || len(result.text) > 200 {
		return nil
	}

	speech := htgotts.Speech{Folder: "audio", Language: target_lang, Handler: &handlers.MPlayer{}}
	filepath, err := speech.CreateSpeechFile(result.text, fmt.Sprintf("voice%d_%d", ctx.EffectiveSender.ChatId, time.Now().UnixNano()))
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, result.text, nil)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	_, err = b.SendVoice(ctx.EffectiveSender.ChatId, gotgbot.InputFileByReader("voice.mp3", file), &gotgbot.SendVoiceOpts{
		Caption: "Voiced version",
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func translate_message(text string, source_lang string, target_lang string, translate_chan chan TranslateResult) {
	result, err := gt.Translate(text, source_lang, target_lang)
	translate_chan <- TranslateResult{text: result, err: err}
}