// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW = "direct_channel_show"
	PREFERENCE_CATEGORY_TUTORIAL_STEPS      = "tutorial_step"
	PREFERENCE_CATEGORY_ADVANCED_SETTINGS   = "advanced_settings"

	PREFERENCE_CATEGORY_DISPLAY_SETTINGS = "display_settings"
	PREFERENCE_NAME_COLLAPSE_SETTING     = "collapse_previews"

	PREFERENCE_CATEGORY_THEME = "theme"
	// the name for theme props is the team id

	PREFERENCE_CATEGORY_LAST     = "last"
	PREFERENCE_NAME_LAST_CHANNEL = "channel"
)

type Preference struct {
	UserId   string `json:"user_id"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

func (o *Preference) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func PreferenceFromJson(data io.Reader) *Preference {
	decoder := json.NewDecoder(data)
	var o Preference
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func (o *Preference) IsValid() *AppError {
	if len(o.UserId) != 26 {
		return NewLocAppError("Preference.IsValid", "model.preference.is_valid.id.app_error", nil, "user_id="+o.UserId)
	}

	if len(o.Category) == 0 || len(o.Category) > 32 {
		return NewLocAppError("Preference.IsValid", "model.preference.is_valid.category.app_error", nil, "category="+o.Category)
	}

	if len(o.Name) > 32 {
		return NewLocAppError("Preference.IsValid", "model.preference.is_valid.name.app_error", nil, "name="+o.Name)
	}

	if utf8.RuneCountInString(o.Value) > 2000 {
		return NewLocAppError("Preference.IsValid", "model.preference.is_valid.value.app_error", nil, "value="+o.Value)
	}

	if o.Category == PREFERENCE_CATEGORY_THEME {
		var unused map[string]string
		if err := json.NewDecoder(strings.NewReader(o.Value)).Decode(&unused); err != nil {
			return NewLocAppError("Preference.IsValid", "model.preference.is_valid.theme.app_error", nil, "value="+o.Value)
		}
	}

	return nil
}

func (o *Preference) PreUpdate() {
	if o.Category == PREFERENCE_CATEGORY_THEME {
		// decode the value of theme (a map of strings to string) and eliminate any invalid values
		var props map[string]string
		if err := json.NewDecoder(strings.NewReader(o.Value)).Decode(&props); err != nil {
			// just continue, the invalid preference value should get caught by IsValid before saving
			return
		}

		colorPattern := regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`)

		// blank out any invalid theme values
		for name, value := range props {
			if name == "image" || name == "type" || name == "codeTheme" {
				continue
			}

			if !colorPattern.MatchString(value) {
				props[name] = "#ffffff"
			}
		}

		if b, err := json.Marshal(props); err == nil {
			o.Value = string(b)
		}
	}
}
