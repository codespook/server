package mongo

import (
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
	"github.com/impactasaurus/server/data"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func (m *mongo) GetMeeting(id string, u auth.User) (impact.Meeting, error) {
	meeting := impact.Meeting{}

	col, closer := m.getMeetingCollection()
	defer closer()

	userOrg, err := u.Organisation()
	if err != nil {
		return meeting, err
	}

	err = col.Find(bson.M{
		"_id":            id,
		"organisationID": userOrg,
	}).One(&meeting)
	if err != nil {
		if mgo.ErrNotFound == err {
			return meeting, data.NewNotFoundError("Meeting")
		}
		return meeting, err
	}
	return meeting, nil
}

type meetingGetter func(col *mgo.Collection, userOrg string) ([]impact.Meeting, error)

func (m *mongo) getMeetings(inner meetingGetter, u auth.User) ([]impact.Meeting, error) {
	col, closer := m.getMeetingCollection()
	defer closer()

	userOrg, err := u.Organisation()
	if err != nil {
		return nil, err
	}
	return inner(col, userOrg)
}

func (m *mongo) GetMeetingsForBeneficiary(beneficiary string, u auth.User) ([]impact.Meeting, error) {
	return m.getMeetings(func(col *mgo.Collection, userOrg string) ([]impact.Meeting, error) {
		results := []impact.Meeting{}
		err := col.Find(bson.M{
			"beneficiary":    beneficiary,
			"organisationID": userOrg,
		}).All(&results)
		return results, err
	}, u)
}

func (m *mongo) GetOSMeetingsForBeneficiary(beneficiary string, outcomeSetID string, u auth.User) ([]impact.Meeting, error) {
	return m.getMeetings(func(col *mgo.Collection, userOrg string) ([]impact.Meeting, error) {
		results := []impact.Meeting{}
		err := col.Find(bson.M{
			"beneficiary":    beneficiary,
			"organisationID": userOrg,
			"outcomeSetID":   outcomeSetID,
		}).All(&results)
		return results, err
	}, u)
}

func (m *mongo) GetOSMeetingsInTimeRange(start, end time.Time, outcomeSetID string, u auth.User) ([]impact.Meeting, error) {
	return m.getMeetings(func(col *mgo.Collection, userOrg string) ([]impact.Meeting, error) {
		results := []impact.Meeting{}
		err := col.Find(bson.M{
			"organisationID": userOrg,
			"outcomeSetID":   outcomeSetID,
			"conducted": bson.M{
				"$gte": start,
				"$lte": end,
			},
		}).All(&results)
		return results, err
	}, u)
}

func (m *mongo) NewMeeting(beneficiaryID, outcomeSetID string, conducted time.Time, u auth.User) (impact.Meeting, error) {
	userOrg, err := u.Organisation()
	if err != nil {
		return impact.Meeting{}, err
	}

	col, closer := m.getMeetingCollection()
	defer closer()

	meeting := impact.Meeting{
		ID:             uuid.NewV4().String(),
		OrganisationID: userOrg,
		OutcomeSetID:   outcomeSetID,
		Beneficiary:    beneficiaryID,
		Conducted:      conducted,
		Created:        time.Now(),
		Modified:       time.Now(),
		User:           u.UserID(),
	}

	if err := col.Insert(meeting); err != nil {
		return impact.Meeting{}, err
	}
	return meeting, nil
}

func (m *mongo) NewAnswer(meetingID string, answer impact.Answer, u auth.User) (impact.Meeting, error) {
	userOrg, err := u.Organisation()
	if err != nil {
		return impact.Meeting{}, err
	}

	col, closer := m.getMeetingCollection()
	defer closer()

	if err := col.Update(bson.M{
		"_id":            meetingID,
		"organisationID": userOrg,
	}, bson.M{
		"$push": bson.M{
			"answers": answer,
		},
	}); err != nil {
		return impact.Meeting{}, err
	}

	return m.GetMeeting(meetingID, u)
}
