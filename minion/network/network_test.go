package network

import (
	"errors"
	"testing"

	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/minion/ovsdb"
	"github.com/quilt/quilt/minion/ovsdb/mocks"
	"github.com/stretchr/testify/mock"
)

func TestRunMaster(t *testing.T) {
	conn := db.New()

	// Supervisor isn't initialized, nothing should happen.
	runMaster(conn)

	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		etcd := view.InsertEtcd()
		etcd.Leader = true
		view.Commit(etcd)

		minion := view.InsertMinion()
		minion.Self = true
		view.Commit(minion)

		label := view.InsertLabel()
		label.Label = "junk"
		view.Commit(label)

		c := view.InsertContainer()
		c.IP = "1.2.3.4"
		view.Commit(c)
		return nil
	})

	anErr := errors.New("err")

	client := new(mocks.Client)
	ovsdb.Open = func() (ovsdb.Client, error) { return nil, anErr }
	runMaster(conn)

	ovsdb.Open = func() (ovsdb.Client, error) {
		return client, nil
	}

	client.On("LogicalSwitchExists", lSwitch).Return(true, nil).Once()
	client.On("Disconnect").Return(nil)
	client.On("ListAddressSets").Return(nil, anErr)
	client.On("ListACLs").Return(nil, anErr)
	client.On("ListSwitchPorts").Return(nil, anErr).Once()

	runMaster(conn)
	client.AssertCalled(t, "Disconnect")

	client.On("LogicalSwitchExists", lSwitch).Return(true, nil).Once()
	client.On("ListSwitchPorts").Return([]ovsdb.SwitchPort{{Name: "1.2.3.5"}}, nil)
	client.On("DeleteSwitchPort", lSwitch, ovsdb.SwitchPort{
		Name: "1.2.3.5", Addresses: nil}).Return(anErr).Once()
	client.On("CreateSwitchPort", lSwitch, "1.2.3.4",
		"02:00:01:02:03:04", "1.2.3.4").Return(anErr).Once()
	runMaster(conn)
	client.AssertCalled(t, "Disconnect")
	client.AssertCalled(t, "ListSwitchPorts")
	client.AssertCalled(t, "DeleteSwitchPort", mock.Anything, mock.Anything)
	client.AssertCalled(t, "CreateSwitchPort", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything)

	client.On("LogicalSwitchExists", lSwitch).Return(true, nil).Once()
	client.On("ListSwitchPorts").Return([]ovsdb.SwitchPort{{Name: "1.2.3.5"}}, nil)
	client.On("DeleteSwitchPort", lSwitch, ovsdb.SwitchPort{
		Name: "1.2.3.5", Addresses: []string(nil)}).Return(nil)
	client.On("CreateSwitchPort", lSwitch, "1.2.3.4",
		"02:00:01:02:03:04", "1.2.3.4").Return(nil).Once()
	runMaster(conn)
	client.AssertCalled(t, "Disconnect")
	client.AssertCalled(t, "ListSwitchPorts")
	client.AssertCalled(t, "DeleteSwitchPort", mock.Anything, mock.Anything)
	client.AssertCalled(t, "CreateSwitchPort", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything)

	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		for _, l := range view.SelectFromLabel(nil) {
			view.Remove(l)
		}

		for _, c := range view.SelectFromContainer(nil) {
			view.Remove(c)
		}

		return nil
	})

	client.On("LogicalSwitchExists", lSwitch).Return(true, nil).Once()
	runMaster(db.New())
	client.AssertNotCalled(t, "CreateLogicalSwitch", mock.Anything)

	client.On("CreateLogicalSwitch", lSwitch).Return(nil)
	client.On("LogicalSwitchExists", lSwitch).Return(false, nil).Once()
	runMaster(conn)
	client.AssertCalled(t, "CreateLogicalSwitch", lSwitch)
}
