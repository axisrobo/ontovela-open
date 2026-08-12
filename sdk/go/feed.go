package ontovela

import "context"

// ConsumeChanges fetches events after the consumer's committed offset and
// commits the last delivered offset atomically from the caller's perspective.
// It returns the events for processing.
func (c *Client) ConsumeChanges(ctx context.Context, consumerID string, limit int, filters ChangeFilter) ([]ChangeEvent, error) {
	cursor, err := c.GetSubscriptionOffset(ctx, consumerID)
	if err != nil {
		return nil, err
	}
	events, err := c.ListChanges(ctx, cursor.CommittedOffset, limit, filters)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return events, nil
	}
	last := events[len(events)-1].Offset
	if _, err := c.CommitSubscriptionOffset(ctx, consumerID, last); err != nil {
		return nil, err
	}
	return events, nil
}
