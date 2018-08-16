package mux

import (
	"github.com/jasonrichardsmith/sentry/sentry"
	"k8s.io/api/admission/v1beta1"
)

type sentryModule struct {
	sentry.Sentry
	ignored []string
}

type SentryMux struct {
	Sentries map[string][]sentryModule
}

func NewFromConfig(c Config) (SentryMux, error) {
	sm := SentryMux{
		Sentries: make(map[string][]sentryModule),
	}
	if c.Limits.Enabled {
		s, err := c.Limits.LoadSentry()
		if err != nil {
			return sm, err
		}
		mod := sentryModule{
			s,
			c.Limits.IgnoredNamespaces,
		}
		sm.Sentries[c.Limits.Type] = []sentryModule{mod}
	}
	if c.Healthz.Enabled {
		s, err := c.Healthz.LoadSentry()
		if err != nil {
			return sm, err
		}
		mod := sentryModule{
			s,
			c.Healthz.IgnoredNamespaces,
		}
		sm.Sentries[c.Healthz.Type] = append(sm.Sentries[c.Healthz.Type], mod)
	}
	if c.Images.Enabled {
		s, err := c.Images.LoadSentry()
		if err != nil {
			return sm, err
		}
		mod := sentryModule{
			s,
			c.Images.IgnoredNamespaces,
		}
		sm.Sentries[c.Images.Type] = append(sm.Sentries[c.Images.Type], mod)
	}
	return sm, nil
}

func (sm sentryModule) Ignore(namespace string) bool {
	for _, ignore := range sm.ignored {
		if ignore == namespace {
			return true

		}
	}
	return false
}

func (sm SentryMux) Admit(receivedAdmissionReview v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	if sms, ok := sm.Sentries[receivedAdmissionReview.Request.Kind.Kind]; ok {
		for _, sm := range sms {
			if !sm.Ignore(receivedAdmissionReview.Request.Namespace) {
				ar := sm.Admit(receivedAdmissionReview)
				if !ar.Allowed {
					return ar
				}
			}
		}

	}
	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true
	return &reviewResponse

}
