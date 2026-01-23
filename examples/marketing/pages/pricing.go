package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// PricingCardProps defines a pricing tier card.
type PricingCardProps struct {
	Name        string
	Price       string
	Period      string
	Description string
	Features    []string
	CTAText     string
	CTAHref     string
	CTAVariant  ui.ButtonVariant
	Highlighted bool
	Badge       string
}

// pricingCard renders a single pricing tier card.
func pricingCard(r *core.Router, props PricingCardProps) core.Node {
	// Build feature list
	featureNodes := make([]core.Node, len(props.Features))
	for i, feature := range props.Features {
		featureNodes[i] = core.Div(core.Class("flex items-center gap-2"),
			core.Span(core.Class("text-green-500"), core.Text("\u2713")), // Checkmark
			core.Span(core.Class("text-gray-600 dark:text-gray-400"), core.Text(feature)),
		)
	}

	// Card styling
	cardClass := "relative"
	if props.Highlighted {
		cardClass = "relative ring-2 ring-blue-500 scale-105 transform"
	}

	// Build card content
	children := []core.Node{}

	// Badge
	if props.Badge != "" {
		children = append(children,
			core.Div(core.Class("absolute -top-3 left-1/2 -translate-x-1/2"),
				core.Span(core.Class("bg-blue-500 text-white text-xs font-semibold px-3 py-1 rounded-full"),
					core.Text(props.Badge)),
			),
		)
	}

	// Card header with name, price, description
	children = append(children,
		ui.CardHeader(ui.CardHeaderProps{
			Class: "text-center",
			Children: []core.Node{
				core.H3(core.Class("text-xl font-bold text-gray-900 dark:text-white"),
					core.Text(props.Name)),
				core.Div(core.Class("mt-4"),
					core.Span(core.Class("text-4xl font-bold text-gray-900 dark:text-white"),
						core.Text(props.Price)),
					core.Span(core.Class("text-gray-500 dark:text-gray-400"),
						core.Text("/"+props.Period)),
				),
				core.P(core.Class("mt-2 text-sm text-gray-500 dark:text-gray-400"),
					core.Text(props.Description)),
			},
		}),
	)

	// Features list
	children = append(children,
		ui.CardContent(ui.CardContentProps{
			Children: []core.Node{
				ui.VStack(ui.StackProps{
					Gap:      "3",
					Children: featureNodes,
				}),
			},
		}),
	)

	// CTA button
	var ctaNode core.Node
	if props.CTAHref != "" {
		// Link button for contact sales
		ctaNode = core.A(
			core.Attrs{
				Href:  props.CTAHref,
				Class: "w-full",
			},
			ui.Button(ui.ButtonProps{
				Variant: props.CTAVariant,
				Class:   "w-full",
				Children: []core.Node{
					core.Text(props.CTAText),
				},
			}),
		)
	} else {
		ctaNode = ui.Button(ui.ButtonProps{
			Variant: props.CTAVariant,
			Class:   "w-full",
			OnClick: func() {
				// Demo: navigate to sign up
				r.Navigate("/contact")
			},
			Children: []core.Node{
				core.Text(props.CTAText),
			},
		})
	}

	children = append(children,
		ui.CardFooter(ui.CardFooterProps{
			Children: []core.Node{ctaNode},
		}),
	)

	return ui.Card(ui.CardProps{
		Class:    cardClass,
		Children: children,
	})
}

// faqItem renders a single FAQ item.
func faqItem(question, answer string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					core.H3(core.Class("font-semibold text-gray-900 dark:text-white mb-2"),
						core.Text(question)),
					core.P(core.Class("text-gray-600 dark:text-gray-400"),
						core.Text(answer)),
				},
			}),
		},
	})
}

// Pricing renders the pricing page with tier cards.
func Pricing(r *core.Router) func() core.Node {
	return func() core.Node {
		// Pricing tiers
		tiers := []PricingCardProps{
			{
				Name:        "Starter",
				Price:       "$9",
				Period:      "month",
				Description: "Perfect for individuals and small projects",
				Features: []string{
					"Up to 3 projects",
					"5GB storage",
					"Basic analytics",
					"Email support",
				},
				CTAText:     "Start Free Trial",
				CTAVariant:  ui.ButtonOutline,
				Highlighted: false,
			},
			{
				Name:        "Pro",
				Price:       "$29",
				Period:      "month",
				Description: "For growing teams that need more",
				Features: []string{
					"Unlimited projects",
					"50GB storage",
					"Advanced analytics",
					"Priority support",
					"API access",
					"Custom domains",
				},
				CTAText:     "Start Free Trial",
				CTAVariant:  ui.ButtonPrimary,
				Highlighted: true,
				Badge:       "Most Popular",
			},
			{
				Name:        "Enterprise",
				Price:       "$99",
				Period:      "month",
				Description: "For organizations with advanced needs",
				Features: []string{
					"Everything in Pro",
					"Unlimited storage",
					"Dedicated support",
					"SSO/SAML",
					"Custom integrations",
					"SLA guarantee",
				},
				CTAText:    "Contact Sales",
				CTAHref:    "/contact",
				CTAVariant: ui.ButtonOutline,
			},
		}

		// Build tier cards
		tierCards := make([]core.Node, len(tiers))
		for i, tier := range tiers {
			capturedTier := tier
			tierCards[i] = pricingCard(r, capturedTier)
		}

		return MarketingLayout(r,
			// Header Section
			Section("Simple, Transparent Pricing",
				"Choose the plan that's right for your team. All plans include a 14-day free trial.",
				ui.Grid(ui.GridProps{
					Cols:  "3",
					Gap:   "8",
					Class: "md:grid-cols-3 grid-cols-1 items-start",
					Children: tierCards,
				}),
			),

			// FAQ Section
			Section("Frequently Asked Questions", "",
				ui.Grid(ui.GridProps{
					Cols:  "3",
					Gap:   "6",
					Class: "md:grid-cols-3 grid-cols-1",
					Children: []core.Node{
						faqItem("Can I change plans later?",
							"Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately."),
						faqItem("What payment methods do you accept?",
							"We accept all major credit cards and PayPal. Enterprise customers can also pay by invoice."),
						faqItem("Is there a free trial?",
							"Yes, all plans include a 14-day free trial. No credit card required to start."),
					},
				}),
			),
		)
	}
}
