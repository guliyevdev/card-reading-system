package smartcard

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jumpycalm/goscard"

	"card-reading-system/internal/state"
)

const DefaultPollInterval = 500 * time.Millisecond

var uidCommand = []byte{0xFF, 0xCA, 0x00, 0x00, 0x00}

type Service struct {
	logger        *log.Logger
	store         *state.Store
	knownReaders  map[string]struct{}
	lastReader    string
	lastUID       string
	lastATR       string
	lastCardKnown bool
}

func NewService(logger *log.Logger, store *state.Store) *Service {
	return &Service{
		logger:       logger,
		store:        store,
		knownReaders: make(map[string]struct{}),
	}
}

func (s *Service) Start(ctx context.Context) {
	if err := goscard.Initialize(noopLogger{}); err != nil {
		s.logger.Printf("PCSC initialize error: %v", err)
		return
	}
	defer goscard.Finalize()

	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		s.pollOnce()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) pollOnce() {
	pcscContext, ret, err := goscard.NewContext(goscard.SCardScopeSystem, nil, nil)
	if err != nil {
		s.logger.Printf("PCSC context error: ret=0x%X err=%v", ret, err)
		s.handleCardRemoval()
		return
	}
	defer func() {
		if releaseRet, releaseErr := pcscContext.Release(); releaseErr != nil {
			s.logger.Printf("PCSC release error: ret=0x%X err=%v", releaseRet, releaseErr)
		}
	}()

	readers, ret, err := pcscContext.ListReaders(nil)
	if err != nil {
		s.logger.Printf("List readers error: ret=0x%X err=%v", ret, err)
		s.handleCardRemoval()
		return
	}

	currentReaders := make(map[string]struct{}, len(readers))
	for _, name := range readers {
		currentReaders[name] = struct{}{}
		if _, ok := s.knownReaders[name]; !ok {
			s.logger.Printf("Reader detected: %s", name)
		}
	}

	for name := range s.knownReaders {
		if _, ok := currentReaders[name]; !ok {
			s.logger.Printf("Reader removed: %s", name)
		}
	}
	s.knownReaders = currentReaders

	presentReaders, atrs, ret, err := pcscContext.ListReadersWithCardPresent(nil)
	if err != nil {
		s.logger.Printf("List readers with card error: ret=0x%X err=%v", ret, err)
		s.handleCardRemoval()
		return
	}

	if len(presentReaders) == 0 {
		s.handleCardRemoval()
		return
	}

	sort.Strings(presentReaders)
	atrByReader := make(map[string]string, len(presentReaders))
	for i, reader := range presentReaders {
		if i < len(atrs) {
			atrByReader[reader] = strings.ToUpper(atrs[i])
		}
	}

	for _, readerName := range presentReaders {
		uid, atr, err := s.readCard(pcscContext, readerName, atrByReader[readerName])
		if err != nil {
			s.logger.Printf("Card processing error (%s): %v", readerName, err)
			continue
		}

		if s.lastCardKnown && s.lastReader == readerName && s.lastUID == uid && s.lastATR == atr {
			return
		}

		s.lastCardKnown = true
		s.lastReader = readerName
		s.lastUID = uid
		s.lastATR = atr

		s.store.Set(state.Card{
			UID:     uid,
			ATR:     atr,
			Reader:  readerName,
			Present: true,
		})

		s.logger.Printf("Card detected. UID=%s, ATR=%s", uid, atr)
		return
	}

	s.handleCardRemoval()
}

func (s *Service) readCard(pcscContext goscard.Context, readerName, atr string) (string, string, error) {
	card, ret, err := pcscContext.Connect(
		readerName,
		goscard.SCardShareShared,
		goscard.SCardProtocolAny,
	)
	if err != nil {
		return "", "", fmt.Errorf("connect ret=0x%X err=%w", ret, err)
	}
	defer func() {
		if disconnectRet, disconnectErr := card.Disconnect(goscard.SCardLeaveCard); disconnectErr != nil {
			s.logger.Printf("Disconnect error (%s): ret=0x%X err=%v", readerName, disconnectRet, disconnectErr)
		}
	}()

	status, ret, err := card.Status()
	if err != nil {
		return "", "", fmt.Errorf("status ret=0x%X err=%w", ret, err)
	}

	ioSendPci, err := protocolPCI(status.ActiveProtocol)
	if err != nil {
		return "", "", err
	}

	response, ret, err := card.Transmit(&ioSendPci, uidCommand, nil)
	if err != nil {
		return "", "", fmt.Errorf("transmit ret=0x%X err=%w", ret, err)
	}

	parsedUID, err := parseUIDResponse(response)
	if err != nil {
		return "", "", err
	}

	if status.Atr != "" {
		atr = strings.ToUpper(status.Atr)
	}

	return parsedUID, atr, nil
}

func (s *Service) handleCardRemoval() {
	if !s.lastCardKnown {
		return
	}

	s.lastCardKnown = false
	s.lastReader = ""
	s.lastUID = ""
	s.lastATR = ""
	s.store.Clear()
	s.logger.Print("Card removed")
}

func protocolPCI(protocol goscard.SCardProtocol) (goscard.SCardIORequest, error) {
	switch protocol {
	case goscard.SCardProtocolT0:
		return goscard.SCardIoRequestT0, nil
	case goscard.SCardProtocolT1:
		return goscard.SCardIoRequestT1, nil
	case goscard.SCardProtocolRaw:
		return goscard.SCardIoRequestRaw, nil
	default:
		return goscard.SCardIORequest{}, fmt.Errorf("unsupported protocol: %s", protocol.String())
	}
}

func parseUIDResponse(response []byte) (string, error) {
	if len(response) < 2 {
		return "", fmt.Errorf("invalid UID response length: %d", len(response))
	}

	status := response[len(response)-2:]
	if !bytes.Equal(status, []byte{0x90, 0x00}) {
		return "", fmt.Errorf("could not get card UID. Status=0x%X", status)
	}

	return strings.ToUpper(hex.EncodeToString(response[:len(response)-2])), nil
}

type noopLogger struct{}

func (noopLogger) Debugf(string, ...interface{}) {}
func (noopLogger) Debug(...interface{})          {}
func (noopLogger) Debugln(...interface{})        {}
func (noopLogger) Infof(string, ...interface{})  {}
func (noopLogger) Info(...interface{})           {}
func (noopLogger) Infoln(...interface{})         {}
func (noopLogger) Warnf(string, ...interface{})  {}
func (noopLogger) Warn(...interface{})           {}
func (noopLogger) Warnln(...interface{})         {}
func (noopLogger) Errorf(string, ...interface{}) {}
func (noopLogger) Error(...interface{})          {}
func (noopLogger) Errorln(...interface{})        {}
