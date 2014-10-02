package ib

import "sync"

type RealtimeBarManager struct {
	AbstractManager
	id  int64
	c   Contract
	bar RealtimeBars
	o   sync.Once
}

func NewRealtimeBarManager(e *Engine, c Contract) (*RealtimeBarManager, error) {
	am, err := NewAbstractManager(e)
	if err != nil {
		return nil, err
	}

	m := &RealtimeBarManager{
		AbstractManager: *am,
		c:               c,
	}

	go m.startMainLoop(m.preLoop, m.receive, m.preDestroy)
	return m, nil
}

func (i *RealtimeBarManager) preLoop() error {
	i.id = i.eng.NextRequestId()
	req := &RequestRealTimeBars{Contract: i.c, WhatToShow: RealTimeTrades, UseRTH: true, BarSize: 5}
	req.SetId(i.id)
	i.eng.Subscribe(i.rc, i.id)
	return i.eng.Send(req)
}

func (i *RealtimeBarManager) preDestroy() {
	i.eng.Unsubscribe(i.rc, i.id)
	req := &CancelRealTimeBars{}
	req.SetId(i.id)
	i.eng.Send(req)
}

func (i *RealtimeBarManager) receive(r Reply) (UpdateStatus, error) {

	switch r.(type) {
	case *ErrorMessage:
		r := r.(*ErrorMessage)
		if r.SeverityWarning() {
			return UpdateFalse, nil
		}
		return UpdateFalse, r.Error()
	case *RealtimeBars:
		r := r.(*RealtimeBars)
		i.bar = *r
		return UpdateTrue, nil
	}

	return UpdateTrue, nil
}

func (rb *RealtimeBarManager) RealtimeBar() RealtimeBars {
	rb.rwm.RLock()
	defer rb.rwm.RUnlock()
	return rb.bar
}
