package tstyle

import (
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/mitchellh/mapstructure"
)

//go:generate go-enum --values --names --flag --nocase --lower

// Style represents the Telegram StyledText
// ENUM(Plain,Unknown,Mention,Hashtag,BotCommand,URL,Email,Bold,Italic,Code,Pre,TextURL,MentionName,Phone,Cashtag,Underline,Strike,BankCard,Spoiler,CustomEmoji,Blockquote)
type Style string

func ParseToStyledText(input map[string]any) (result *message.StyledTextOption, err error) {
	style, err := ParseStyle(input["style"].(string))
	if err != nil {
		return nil, errors.Wrap(err, "parse style")
	}
	switch style {
	case StylePre:
		o := new(preStyle)
		if err = mapstructure.WeakDecode(input, &o); err != nil {
			return nil, errors.Wrap(err, "decode options")
		}
		r := styling.Pre(o.Text, o.Language)
		result = &r
		return result, err
	case StyleTextURL:
		o := new(textURLStyle)
		if err = mapstructure.WeakDecode(input, &o); err != nil {
			return nil, errors.Wrap(err, "decode options")
		}
		r := styling.TextURL(o.Text, o.URL)
		result = &r
		return result, err
	case StyleMentionName:
		return nil, errors.New("unsupported style")
	case StyleCustomEmoji:
		o := new(customEmojiStyle)
		if err = mapstructure.WeakDecode(input, &o); err != nil {
			return nil, errors.Wrap(err, "decode options")
		}
		r := styling.CustomEmoji(o.Text, o.DocumentID)
		result = &r
		return result, err
	default:
		o := new(commonStyle)
		if err = mapstructure.WeakDecode(input, &o); err != nil {
			return nil, errors.Wrap(err, "decode options")
		}
		var r message.StyledTextOption
		r, err = processCommonStyle(*o)
		if err != nil {
			return nil, errors.Wrap(err, "process common style")
		}
		result = &r
		return result, err
	}
}

func processCommonStyle(commonStyle commonStyle) (result message.StyledTextOption, err error) {
	style, err := ParseStyle(commonStyle.Style)
	if err != nil {
		err = errors.Wrap(err, "parse style")
		return result, err
	}
	switch style {
	case StylePlain:
		result = styling.Plain(commonStyle.Text)
	case StyleUnknown:
		result = styling.Unknown(commonStyle.Text)
	case StyleMention:
		result = styling.Mention(commonStyle.Text)
	case StyleHashtag:
		result = styling.Hashtag(commonStyle.Text)
	case StyleBotCommand:
		result = styling.BotCommand(commonStyle.Text)
	case StyleURL:
		result = styling.URL(commonStyle.Text)
	case StyleEmail:
		result = styling.Email(commonStyle.Text)
	case StyleBold:
		result = styling.Bold(commonStyle.Text)
	case StyleItalic:
		result = styling.Italic(commonStyle.Text)
	case StyleCode:
		result = styling.Code(commonStyle.Text)
	case StylePre:
		err = errors.New("special style in common style switch")
	case StyleTextURL:
		err = errors.New("special style in common style switch")
	case StyleMentionName:
		err = errors.New("special style in common style switch")
	case StylePhone:
		result = styling.Phone(commonStyle.Text)
	case StyleCashtag:
		result = styling.Cashtag(commonStyle.Text)
	case StyleUnderline:
		result = styling.Underline(commonStyle.Text)
	case StyleStrike:
		result = styling.Strike(commonStyle.Text)
	case StyleBankCard:
		result = styling.BankCard(commonStyle.Text)
	case StyleSpoiler:
		result = styling.Spoiler(commonStyle.Text)
	case StyleCustomEmoji:
		err = errors.New("special style in common style switch")
	case StyleBlockquote:
		result = styling.Blockquote(commonStyle.Text, false)
	default:
		err = errors.Wrap(ErrInvalidStyle, "switch style")
	}
	return result, err
}

type commonStyle struct {
	Style string
	Text  string
}

type preStyle struct {
	Style    string
	Text     string
	Language string
}

type textURLStyle struct {
	Style string
	Text  string
	URL   string
}

type customEmojiStyle struct {
	Style      string
	Text       string
	DocumentID int64
}
