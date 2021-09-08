package fixtures

import (
	"embed"
	"fmt"
	"sort"

	"github.com/spf13/afero"
)

//go:embed triggers/*
var triggers embed.FS

// Events is a mapping of pre-built trigger events and the corresponding json file
var Events = map[string]string{
	"account.updated":                          "triggers/account.updated.json",
	"balance.available":                        "triggers/balance.available.json",
	"charge.captured":                          "triggers/charge.captured.json",
	"charge.dispute.created":                   "triggers/charge.disputed.created.json",
	"charge.failed":                            "triggers/charge.failed.json",
	"charge.refunded":                          "triggers/charge.refunded.json",
	"charge.succeeded":                         "triggers/charge.succeeded.json",
	"checkout.session.async_payment_failed":    "triggers/checkout.session.async_payment_failed.json",
	"checkout.session.async_payment_succeeded": "triggers/checkout.session.async_payment_succeeded.json",
	"checkout.session.completed":               "triggers/checkout.session.completed.json",
	"customer.created":                         "triggers/customer.created.json",
	"customer.deleted":                         "triggers/customer.deleted.json",
	"customer.updated":                         "triggers/customer.updated.json",
	"customer.source.created":                  "triggers/customer.source.created.json",
	"customer.source.updated":                  "triggers/customer.source.updated.json",
	"customer.subscription.created":            "triggers/customer.subscription.created.json",
	"customer.subscription.deleted":            "triggers/customer.subscription.deleted.json",
	"customer.subscription.updated":            "triggers/customer.subscription.updated.json",
	"invoice.created":                          "triggers/invoice.created.json",
	"invoice.finalized":                        "triggers/invoice.finalized.json",
	"invoice.payment_action_required":          "triggers/invoice.payment_action_required.json",
	"invoice.payment_failed":                   "triggers/invoice.payment_failed.json",
	"invoice.payment_succeeded":                "triggers/invoice.payment_succeeded.json",
	"invoice.updated":                          "triggers/invoice.updated.json",
	"issuing_authorization.request":            "triggers/issuing_authorization.request.json",
	"issuing_card.created":                     "triggers/issuing_card.created.json",
	"issuing_cardholder.created":               "triggers/issuing_cardholder.created.json",
	"payment_intent.amount_capturable_updated": "triggers/payment_intent.amount_capturable_updated.json",
	"payment_intent.created":                   "triggers/payment_intent.created.json",
	"payment_intent.payment_failed":            "triggers/payment_intent.payment_failed.json",
	"payment_intent.succeeded":                 "triggers/payment_intent.succeeded.json",
	"payment_intent.canceled":                  "triggers/payment_intent.canceled.json",
	"payment_method.attached":                  "triggers/payment_method.attached.json",
	"payout.created":                           "triggers/payout.created.json",
	"payout.updated":                           "triggers/payout.updated.json",
	"plan.created":                             "triggers/plan.created.json",
	"plan.deleted":                             "triggers/plan.deleted.json",
	"plan.updated":                             "triggers/plan.updated.json",
	"product.created":                          "triggers/product.created.json",
	"product.deleted":                          "triggers/product.deleted.json",
	"product.updated":                          "triggers/product.updated.json",
	"setup_intent.canceled":                    "triggers/setup_intent.canceled.json",
	"setup_intent.created":                     "triggers/setup_intent.created.json",
	"setup_intent.setup_failed":                "triggers/setup_intent.setup_failed.json",
	"setup_intent.succeeded":                   "triggers/setup_intent.succeeded.json",
	"subscription_schedule.canceled":           "triggers/subscription_schedule.canceled.json",
	"subscription_schedule.created":            "triggers/subscription_schedule.created.json",
	"subscription_schedule.released":           "triggers/subscription_schedule.released.json",
	"subscription_schedule.updated":            "triggers/subscription_schedule.updated.json",
	"quote.created":                            "triggers/quote.created.json",
	"quote.canceled":                           "triggers/quote.canceled.json",
	"quote.finalized":                          "triggers/quote.finalized.json",
	"quote.accepted":                           "triggers/quote.accepted.json",
}

// BuildFromFixture creates a new fixture struct for a file
func BuildFromFixture(fs afero.Fs, apiKey string, stripeAccount string, skip []string, overrides []string, additions []string, removals []string, apiBaseURL string, jsonFile string) (*Fixture, error) {
	fixture, err := NewFixture(
		fs,
		apiKey,
		stripeAccount,
		skip,
		apiBaseURL,
		jsonFile,
	)
	if err != nil {
		return nil, err
	}

	if len(overrides) != 0 {
		fixture.Override(overrides)
	}

	if len(additions) != 0 {
		fixture.Add(additions)
	}

	if len(removals) != 0 {
		fixture.Remove(removals)
	}

	return fixture, nil
}

// EventList prints out a padded list of supported trigger events for printing the help file
func EventList() string {
	var eventList string
	for _, event := range EventNames() {
		eventList += fmt.Sprintf("  %s\n", event)
	}

	return eventList
}

// EventNames returns an array of all the event names
func EventNames() []string {
	names := []string{}
	for name := range Events {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// Trigger triggers a Stripe event.
func Trigger(event string, stripeAccount string, skip []string, overrides []string, additions []string, removals []string, baseURL string, apiKey string) ([]string, error) {
	fs := afero.NewOsFs()

	var fixture *Fixture
	var err error

	if file, ok := Events[event]; ok {
		fixture, err = BuildFromFixture(fs, apiKey, stripeAccount, skip, overrides, additions, removals, baseURL, file)
		if err != nil {
			return nil, err
		}
	} else {
		exists, _ := afero.Exists(fs, event)
		if !exists {
			return nil, fmt.Errorf(fmt.Sprintf("The event ‘%s’ is not supported by the Stripe CLI.", event))
		}

		fixture, err = BuildFromFixture(fs, apiKey, stripeAccount, skip, overrides, additions, removals, baseURL, event)
		if err != nil {
			return nil, err
		}
	}

	requestNames, err := fixture.Execute()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Trigger failed: %s\n", err))
	}

	return requestNames, nil
}

func reverseMap() map[string]string {
	reversed := make(map[string]string)
	for name, file := range Events {
		reversed[file] = name
	}

	return reversed
}
