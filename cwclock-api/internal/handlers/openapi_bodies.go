package handlers

import (
	"reflect"

	"cwclock-api/internal/models"
	"cwclock-api/internal/openapi"
)

// EndpointDocs maps "METHOD /route" to the request shape the OpenAPI generator
// can't derive from the router alone: the JSON body struct (its schema is
// reflected, so it tracks the struct), and the query parameters the handler
// reads. Only this route→doc mapping is maintained by hand — add a line to
// document another endpoint's body or query params.
//
// It lives in this package (not openapi) because almost every payload is an
// unexported handler type (registerPayload, clientPayload, chatPayload, …)
// that only this package can name.
//
// The route keys are the patterns chi.Walk emits, so an index route registered
// as Post("/") carries a trailing slash (e.g. "POST /v1/organizations/") while
// a named sub-path does not.
func EndpointDocs() map[string]openapi.EndpointDoc {
	t := reflect.TypeOf
	body := func(v any) openapi.EndpointDoc { return openapi.EndpointDoc{Body: t(v)} }
	// contract documents a route's whole exchange — the request body, the 200
	// answer and the refusal — each with a sentence of its own. Only an endpoint
	// written to somebody else's contract wants this: whoever integrates against
	// it reads /openapi.json and never this repository, so "OK" tells them
	// nothing about the fields their parser has to read. Every other route stays
	// on body(...) and the bare 200, because its reply is readable in the handler
	// and a response schema maintained by hand would drift from it.
	contract := func(in, out, bad any, success, refusal string) openapi.EndpointDoc {
		return openapi.EndpointDoc{
			Body:               t(in),
			Answer:             t(out),
			AnswerDescription:  success,
			Failure:            t(bad),
			FailureDescription: refusal,
		}
	}
	q := func(query ...openapi.Param) openapi.EndpointDoc { return openapi.EndpointDoc{Query: query} }
	p := func(name string) openapi.Param { return openapi.Param{Name: name} }
	ints := func(name string) openapi.Param { return openapi.Param{Name: name, Type: "integer"} }

	return map[string]openapi.EndpointDoc{
		// --- public (registration / login / password / contact) ---
		"POST /v1/contact":               body(contactPayload{}),
		"POST /v1/users/":                body(registerPayload{}),
		"POST /v1/users/login":           body(credentials{}),
		"POST /v1/users/forgot-password": body(forgotPasswordPayload{}),
		"POST /v1/users/reset-password":  body(resetPasswordPayload{}),

		// --- login MFA (challenge/ceremony tokens, pre-session) ---
		"POST /v1/users/login/mfa/totp":            body(mfaTOTPLoginPayload{}),
		"POST /v1/users/login/mfa/webauthn/begin":  body(mfaChallengeTokenPayload{}),
		"POST /v1/users/login/mfa/webauthn/finish": body(mfaWebAuthnFinishPayload{}),

		// --- own account (profile / config file / MFA enrolment) ---
		"PUT /v1/users/me":                               body(updateProfilePayload{}),
		"PUT /v1/users/me/picture":                       body(updatePicturePayload{}),
		"POST /v1/users/me/config/file":                  body(configPayload{}),
		"POST /v1/users/me/config/qr":                    body(configPayload{}),
		"POST /v1/users/me/mfa/totp/confirm":             body(mfaCodePayload{}),
		"POST /v1/users/me/mfa/webauthn/register/finish": body(mfaWebAuthnRegisterFinishPayload{}),

		// --- admin (superuser) ---
		"PUT /v1/admin/users/{id}": body(adminUpdateUserPayload{}),
		"PUT /v1/admin/ip-bans/":   body(models.IpBanAction{}),

		// --- AI assistant ---
		"POST /v1/ai/connections/":            body(aiConnectionPayload{}),
		"PUT /v1/ai/connections/{id}":         body(aiConnectionPayload{}),
		"POST /v1/ai/connections/order":       body(aiOrderPayload{}),
		"POST /v1/ai/connections/{id}/shares": body(aiGrantPayload{}),
		"POST /v1/ai/chat":                    body(chatPayload{}),
		// The CWAI external adapter, the one endpoint whose reply is a contract
		// somebody else parses: https://www.cwcloud.tech/docs/tutorials/cwai#external-adapter-interface
		"POST /v1/ai/cwcloud/adapter": contract(cwcloudAdapterPayload{}, cwcloudAdapterResponse{}, errorBody{},
			"The answer. `status` is \"ok\", the assistant's reply is the text in `message` (Markdown, and the only place it is — nothing is attached), "+
				"and `usage` carries what the call cost on the subscription that was spent: `prompt_tokens`, `completion_tokens` and their sum `total_tokens`.",
			"A refusal, carrying no answer: `message` says what went wrong and `i18n_code` is cwclock's own code for it. "+
				"The HTTP status says it too — 400 a malformed body or an empty prompt, 401 a bad or missing credential, 404 an `adapter` naming a subscription that is not the caller's, "+
				"428 a caller with no AI subscription at all, 502 a provider that refused or could not be reached."),
		"DELETE /v1/ai/conversations/": body(deleteConversationsPayload{}),

		// --- organizations / members ---
		"POST /v1/organizations/":                              body(organizationPayload{}),
		"PATCH /v1/organizations/{orgId}/":                     body(organizationPayload{}),
		"PATCH /v1/organizations/{orgId}/external-connections": body(models.ExternalConnection{}),
		"PUT /v1/organizations/{orgId}/owner":                  body(transferOwnershipPayload{}),
		"POST /v1/organizations/{orgId}/members/":              body(memberPayload{}),
		"PUT /v1/organizations/{orgId}/members/{userId}":       body(memberPayload{}),
		"PUT /v1/organizations/{orgId}/members/{userId}/rate":  body(memberRatePayload{}),

		// --- clients / projects ---
		"POST /v1/organizations/{orgId}/clients/":                     body(clientPayload{}),
		"PUT /v1/organizations/{orgId}/clients/{clientId}":            body(clientPayload{}),
		"PUT /v1/organizations/{orgId}/clients/{clientId}/transfer":   body(transferClientPayload{}),
		"POST /v1/organizations/{orgId}/clients/{clientId}/projects/": body(projectPayload{}),
		"PUT /v1/organizations/{orgId}/projects/{projectId}":          body(projectPayload{}),

		// --- time entries ---
		"POST /v1/organizations/{orgId}/time-entries/":     body(timeEntryPayload{}),
		"PUT /v1/organizations/{orgId}/time-entries/{id}":  body(timeEntryPayload{}),
		"POST /v1/organizations/{orgId}/time-entries/fill": body(fillPayload{}),
		"POST /v1/admin/holidays/":                         body(holidayPayload{}),
		"PUT /v1/admin/holidays/{id}":                      body(holidayPayload{}),

		// --- reports (POST with a filter body, streamed PDF/CSV or JSON) ---
		"POST /v1/organizations/{orgId}/reports/detailed": body(exportRequest{}),
		"POST /v1/organizations/{orgId}/reports/summary":  body(exportRequest{}),

		// --- invoices ---
		"POST /v1/organizations/{orgId}/invoices/preview":    body(invoiceRequest{}),
		"POST /v1/organizations/{orgId}/invoices/":           body(invoiceRequest{}),
		"PUT /v1/organizations/{orgId}/invoices/{invoiceId}": body(updateInvoiceStatusPayload{}),

		// --- export jobs ---
		"POST /v1/organizations/{orgId}/export-jobs/":       body(exportJobPayload{}),
		"PUT /v1/organizations/{orgId}/export-jobs/{jobId}": body(exportJobPayload{}),

		// --- GET query params (main list / search endpoints) ---
		"GET /v1/fields":               q(p("country")),
		"GET /v1/user/confirmation":    q(openapi.Param{Name: "token", Required: true}),
		"GET /v1/users/search":         q(p("q")),
		"GET /v1/admin/users/":         q(p("term"), p("role"), ints("startIndex"), ints("maxResults")),
		"GET /v1/admin/organizations/": q(p("term"), ints("startIndex"), ints("maxResults")),
		"GET /v1/admin/ip-bans/":       q(p("term"), ints("page")),
		"GET /v1/admin/audit/":         q(ints("page"), ints("size"), p("action"), p("subjectType"), p("subject")),
		"GET /v1/admin/holidays/":      q(p("country")),
		"GET /v1/organizations/{orgId}/clients/{clientId}/working-days": q(openapi.Param{Name: "from", Required: true}, openapi.Param{Name: "to", Required: true}, p("projectId")),
		"GET /v1/organizations/{orgId}/projects/":                       q(p("clientId")),
		"GET /v1/organizations/{orgId}/time-entries/":                   q(ints("page"), ints("pageSize")),
		"GET /v1/organizations/{orgId}/time-entries/range":              q(openapi.Param{Name: "start", Required: true}, openapi.Param{Name: "end", Required: true}),
		"GET /v1/organizations/{orgId}/invoices/":                       q(p("clientId"), p("start"), p("end")),
	}
}
