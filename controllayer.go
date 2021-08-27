package touch

type ControlStateMask int

const (
	ControlStateNormal      ControlStateMask = 0
	ControlStateHighlighted ControlStateMask = 1 << iota
	ControlStateDisabled
)

type ControlAction int

const (
	ControlTapped ControlAction = iota
	ControlActionsCount
)

type ControlLayer struct {
	BasicLayer
	State ControlStateMask
}

type ControlDelegate interface {
	StateDidChange()
	HandleAction(ControlAction)
}

func (c *ControlLayer) SetState(state ControlStateMask) {
	if state != c.State {
		c.State = state
		if del, ok := c.Self.(ControlDelegate); ok {
			del.StateDidChange()
		}
	}
}

func (c *ControlLayer) TriggerAction(action ControlAction) {
	if del, ok := c.Self.(ControlDelegate); ok {
		del.HandleAction(action)
	}
}

/* State Getters and Setters */

func (c *ControlLayer) IsHighlighted() bool {
	return c.State&ControlStateHighlighted == ControlStateHighlighted
}

func (c *ControlLayer) SetHighlighted(highlighted bool) {
	if highlighted {
		c.SetState(c.State | ControlStateHighlighted)
	} else {
		c.SetState(c.State &^ ControlStateHighlighted)
	}
}

func (c *ControlLayer) IsDisabled() bool {
	return c.State&ControlStateDisabled == ControlStateDisabled
}

func (c *ControlLayer) SetDisabled(disabled bool) {
	if disabled {
		c.SetState(c.State | ControlStateDisabled)
	} else {
		c.SetState(c.State &^ ControlStateDisabled)
	}
}

/* Touch Event Handling; Button behavior is the default. */

func (c *ControlLayer) StartTouch(event TouchEvent) {
	c.SetHighlighted(event.In(c.Rectangle))
}

func (c *ControlLayer) UpdateTouch(event TouchEvent) {
	c.SetHighlighted(event.In(c.Rectangle))
}

func (c *ControlLayer) EndTouch(event TouchEvent) {
	if event.In(c.Rectangle) {
		c.TriggerAction(ControlTapped)
	}
	c.SetHighlighted(false)
}
