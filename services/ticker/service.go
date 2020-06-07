package timer

import (
	"strconv"
	"time"

	app "timer-atomic-clock/app/interface"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	app    app.AppImpl
	ticker *time.Ticker
}

func CreateService(a app.AppImpl) *Service {

	// Preparing service
	service := &Service{
		app: a,
	}

	return service
}

func (service *Service) RunTickerCluster() {
	log.Info("Stanby ...")
	timer := time.AfterFunc(1100*time.Millisecond,
		func() {
			log.Info("Is Master.")
			service.StartTicker(100)
		},
	)

	//subscribe queue
	signalBus := service.app.GetSignalBus()
	signalBus.Subscribe(
		"timer.ticker",
		func(m *nats.Msg) {
			//log.Info(string(m.Data))
			timer.Reset(1100 * time.Millisecond)
		})

}

func (service *Service) StartTicker(duration time.Duration) {

	// Start ticker
	service.ticker = time.NewTicker(duration * time.Millisecond)
	defer service.ticker.Stop()

	var old int64 = 0
	for {
		select {
		case <-service.ticker.C:
			now := time.Now().Unix()
			if now == old || now < old {
				continue
			}

			//Publish to queue
			signalBus := service.app.GetSignalBus()
			signalBus.Emit("timer.ticker", []byte(strconv.FormatInt(now, 10)))
			old = now
		}
	}

}

func (service *Service) StopTicker() {

	// Stop timer
	service.ticker.Stop()

}
