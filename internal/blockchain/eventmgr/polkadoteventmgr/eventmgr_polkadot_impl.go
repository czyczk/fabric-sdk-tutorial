package polkadoteventmgr

import (
	"net/http"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
	log "github.com/sirupsen/logrus"
)

type PolkadotEventManager struct {
	eventmgr.EventManagerBase
	ctx            *chaincodectx.PolkadotChaincodeCtx
	client         *http.Client
	mapLock        sync.RWMutex
	updateInterval time.Duration
}

func NewPolkadotEventManager(ctx *chaincodectx.PolkadotChaincodeCtx) *PolkadotEventManager {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &PolkadotEventManager{
		EventManagerBase: eventmgr.EventManagerBase{
			QuitChanMap: make(map[eventmgr.IEventRegistration]chan struct{}),
		},
		ctx:            ctx,
		client:         client,
		mapLock:        sync.RWMutex{},
		updateInterval: 1000 * time.Millisecond,
	}
}

func (m *PolkadotEventManager) RegisterEvent(eventID string) (eventmgr.IEventRegistration, <-chan eventmgr.IEvent, error) {
	// Encapsulate the contractAddress in a `chaincodectx.PolkadotChaincodeCtx` object.
	polkadotReg := &PolkadotEventRegistration{
		contractAddress: m.ctx.ContractAddress,
		eventID:         eventID,
	}

	err := registerPolkadotEvent(m.ctx, m.client, polkadotReg.contractAddress, polkadotReg.eventID)
	if err != nil {
		log.Error(err)
	}

	notifier := make(chan eventmgr.IEvent)
	quitChan := make(chan struct{})
	// Background task: send HTTP requests to retrieve events and wrap received Polkadot events to `eventmgr.IEvent` objects.
	go func() {
		for {
			select {
			case <-quitChan:
				close(notifier)
				return
			default:
				polkadotEvents, err := releasePolkadotEvents(m.ctx, m.client, polkadotReg)
				if err != nil {
					log.Error(err)
				}

				for _, polkadotEvent := range polkadotEvents {
					notifier <- polkadotEvent
				}

				time.Sleep(m.updateInterval)
			}
		}
	}()

	// Record the registration. The quit handles are useful to stop the listening processes in the background.
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	m.QuitChanMap[polkadotReg] = quitChan

	return polkadotReg, notifier, nil
}

func (m *PolkadotEventManager) UnregisterEvent(reg eventmgr.IEventRegistration) error {
	polkadotReg := reg.(*PolkadotEventRegistration)

	// Use the quit chan to stop the corresponding background process for event conversion
	m.mapLock.RLock()
	quitChan := m.QuitChanMap[polkadotReg]
	quitChan <- struct{}{}
	m.mapLock.RUnlock()

	// Now the quit chan entry is not useful. Close the quit chan and remove it from the map.
	close(quitChan)
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	delete(m.QuitChanMap, polkadotReg)

	return nil
}
