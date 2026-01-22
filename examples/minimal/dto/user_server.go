//go:build !js || !wasm

package dto

import (
	"github.com/dougbarrett/gux/examples/minimal/models"
)

// FromModel converts a User model to UserDetail DTO.
// Implements core.DTOMapper interface.
// This file is only compiled for server-side builds (not WASM).
func (d *UserDetail) FromModel(model interface{}) interface{} {
	user, ok := model.(models.User)
	if !ok {
		if userPtr, ok := model.(*models.User); ok {
			user = *userPtr
		} else {
			return d
		}
	}

	result := UserDetail{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Convert posts to PostBrief
	if len(user.Posts) > 0 {
		result.Posts = make([]PostBrief, len(user.Posts))
		for i, post := range user.Posts {
			result.Posts[i] = PostBrief{
				ID:        post.ID,
				Title:     post.Title,
				CreatedAt: post.CreatedAt,
			}
		}
	}

	return result
}
