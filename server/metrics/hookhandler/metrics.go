package hookhandler

import "github.com/prometheus/client_golang/prometheus"

var (
	HookRecieves = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_recieved_hooks",
		Help: "hooks recieved and processed by hookhandler",
		// vcs_type: bitbucket | github | etc
		// event_type: pullrequest | push
	}, []string{"vcs_type", "event_type"})
	FailedToTellWerker = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_hookhandler_failed_werker_tell",
		Help: "hookhandler failed to add job to werker queue",
	})
	UnprocessibleEvent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_hookhandler_unprocessible_event_type",
		Help: "hookhandler unable to process this type of event",
	}, []string{"event", "vcstype"},
	)
	FailedTranslation = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_hookhandler_failed_translation",
		Help: "hookhandler failed to translate incoming hook to pr/push object",
		// event_type: pullrequest | push
	}, []string{"event_type"})
)

func init() {
	prometheus.MustRegister(HookRecieves)
}
