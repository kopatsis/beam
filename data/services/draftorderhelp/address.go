package draftorderhelp

import "beam/data/models"

func MergeAddresses(draft *models.DraftOrder, contacts []*models.Contact) bool {

	mod := false

	for _, ct := range contacts {
		found := false
		for _, c := range draft.ListedContacts {
			if c.ID == ct.ID {
				found = true
				break
			}
		}
		if !found {
			draft.ListedContacts = append(draft.ListedContacts, ct)
			mod = true
		}
	}

	var updatedListedContacts []*models.Contact

	for _, c := range draft.ListedContacts {
		if c.ID != 0 {
			found := false
			for _, ct := range contacts {
				if c.ID == ct.ID {
					found = true
					break
				}
			}
			if found {
				updatedListedContacts = append(updatedListedContacts, c)
			} else {
				mod = true
			}
		}
	}

	draft.ListedContacts = updatedListedContacts

	if draft.ShippingContact == nil {
		if len(draft.ListedContacts) > 0 {
			draft.ShippingContact = draft.ListedContacts[0]
			mod = true
		}
	} else if draft.ShippingContact.ID != 0 {
		found := false
		for _, ct := range contacts {
			if draft.ShippingContact.ID == ct.ID {
				found = true
				break
			}
		}
		if !found {
			if len(draft.ListedContacts) > 0 {
				draft.ShippingContact = draft.ListedContacts[0]
			} else {
				draft.ShippingContact = nil
			}
			mod = true
		}
	}

	return mod
}
