package reply

import (
	"context"
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"go.fm/pkg/emojis"
	"time"
)

type ResponseManager struct {
	state       *state.State
	interaction *discord.InteractionEvent
	token       string
	appID       discord.AppID
	deferred    bool
	responded   bool
}

func New(s *state.State, i *discord.InteractionEvent) *ResponseManager {
	return &ResponseManager{
		state:       s,
		interaction: i,
		token:       i.Token,
		appID:       i.AppID,
	}
}

type ResponseBuilder struct {
	manager *ResponseManager
	data    api.InteractionResponseData
}

func (rm *ResponseManager) Reply() *ResponseBuilder {
	return &ResponseBuilder{
		manager: rm,
		data:    api.InteractionResponseData{},
	}
}

func (rm *ResponseManager) Defer(flags ...discord.MessageFlags) *DeferredResponse {
	if rm.responded {
		return &DeferredResponse{manager: rm, err: fmt.Errorf("already responded")}
	}

	var combinedFlags discord.MessageFlags
	for _, flag := range flags {
		combinedFlags |= flag
	}

	response := api.InteractionResponse{Type: api.DeferredMessageInteractionWithSource}
	if combinedFlags != 0 {
		response.Data = &api.InteractionResponseData{Flags: combinedFlags}
	}

	err := rm.state.RespondInteraction(rm.interaction.ID, rm.token, response)
	rm.deferred = true
	rm.responded = true

	return &DeferredResponse{manager: rm, err: err}
}

func (rm *ResponseManager) FollowUp() *FollowUpBuilder {
	return &FollowUpBuilder{manager: rm, data: api.InteractionResponseData{}}
}

func (rb *ResponseBuilder) Content(content string) *ResponseBuilder {
	rb.data.Content = option.NewNullableString(content)
	return rb
}

func (rb *ResponseBuilder) Embed(embed discord.Embed) *ResponseBuilder {
	rb.data.Embeds = &[]discord.Embed{embed}
	return rb
}

func (rb *ResponseBuilder) Embeds(embeds ...discord.Embed) *ResponseBuilder {
	rb.data.Embeds = &embeds
	return rb
}

func (rb *ResponseBuilder) Components(components discord.ContainerComponents) *ResponseBuilder {
	rb.data.Components = &components
	return rb
}

func (rb *ResponseBuilder) Flags(flags ...discord.MessageFlags) *ResponseBuilder {
	for _, flag := range flags {
		rb.data.Flags |= flag
	}
	return rb
}

func (rb *ResponseBuilder) Send() error {
	if rb.manager.responded {
		return fmt.Errorf("already responded")
	}

	err := rb.manager.state.RespondInteraction(
		rb.manager.interaction.ID,
		rb.manager.token,
		api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &rb.data,
		},
	)

	rb.manager.responded = true
	return err
}

type DeferredResponse struct {
	manager *ResponseManager
	err     error
}

func (dr *DeferredResponse) Error() error {
	return dr.err
}

func (dr *DeferredResponse) Edit() *EditBuilder {
	return &EditBuilder{manager: dr.manager, data: api.EditInteractionResponseData{}}
}

type EditBuilder struct {
	manager *ResponseManager
	data    api.EditInteractionResponseData
}

func (eb *EditBuilder) Content(content string) *EditBuilder {
	eb.data.Content = option.NewNullableString(content)
	return eb
}

func (eb *EditBuilder) Embed(embed discord.Embed) *EditBuilder {
	eb.data.Embeds = &[]discord.Embed{embed}
	return eb
}

func (eb *EditBuilder) Embeds(embeds ...discord.Embed) *EditBuilder {
	eb.data.Embeds = &embeds
	return eb
}

func (eb *EditBuilder) Components(components discord.ContainerComponents) *EditBuilder {
	eb.data.Components = &components
	return eb
}

func (eb *EditBuilder) Clear() *EditBuilder {
	eb.data.Content = option.NewNullableString("")
	eb.data.Embeds = &[]discord.Embed{}
	eb.data.Components = &discord.ContainerComponents{}
	return eb
}

func (eb *EditBuilder) Send() (*discord.Message, error) {
	return eb.manager.state.EditInteractionResponse(eb.manager.appID, eb.manager.token, eb.data)
}

type FollowUpBuilder struct {
	manager *ResponseManager
	data    api.InteractionResponseData
}

func (fb *FollowUpBuilder) Content(content string) *FollowUpBuilder {
	fb.data.Content = option.NewNullableString(content)
	return fb
}

func (fb *FollowUpBuilder) Embed(embed discord.Embed) *FollowUpBuilder {
	fb.data.Embeds = &[]discord.Embed{embed}
	return fb
}

func (fb *FollowUpBuilder) Embeds(embeds ...discord.Embed) *FollowUpBuilder {
	fb.data.Embeds = &embeds
	return fb
}

func (fb *FollowUpBuilder) Components(components discord.ContainerComponents) *FollowUpBuilder {
	fb.data.Components = &components
	return fb
}

func (fb *FollowUpBuilder) Flags(flags ...discord.MessageFlags) *FollowUpBuilder {
	for _, flag := range flags {
		fb.data.Flags |= flag
	}
	return fb
}

func (fb *FollowUpBuilder) Send() (*discord.Message, error) {
	return fb.manager.state.FollowUpInteraction(fb.manager.appID, fb.manager.token, fb.data)
}

func (rm *ResponseManager) Quick(content string, flags ...discord.MessageFlags) error {
	builder := rm.Reply().Content(content)
	if len(flags) > 0 {
		builder = builder.Flags(flags...)
	}
	return builder.Send()
}

func (rm *ResponseManager) QuickEmbed(embed discord.Embed, flags ...discord.MessageFlags) error {
	builder := rm.Reply().Embed(embed)
	if len(flags) > 0 {
		builder = builder.Flags(flags...)
	}
	return builder.Send()
}

func (rm *ResponseManager) AutoDefer(fn func(edit *EditBuilder) error, flags ...discord.MessageFlags) error {
	deferred := rm.Defer(flags...)
	if deferred.Error() != nil {
		return deferred.Error()
	}

	editBuilder := deferred.Edit()
	err := fn(editBuilder)
	if err != nil {
		_, err := editBuilder.Clear().Embed(ErrorEmbed(err.Error())).Send()
		return err
	}

	return nil
}

func ErrorEmbed(description string) discord.Embed {
	return discord.Embed{
		Description: fmt.Sprintf("%s %s", emojis.EmojiCross, description),
		Color:       0xFF0000,
	}
}

func SuccessEmbed(description string) discord.Embed {
	return discord.Embed{
		Description: fmt.Sprintf("%s %s", emojis.EmojiCheck, description),
		Color:       0x00FF00,
	}
}

func InfoEmbed(description string) discord.Embed {
	return discord.Embed{
		Description: fmt.Sprintf("%s %s", emojis.EmojiChat, description),
		Color:       0x0099FF,
	}
}

func WithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
