package checkpoint

import (
	"testing"
	"fmt"

	"nimo-shake/common"

	"github.com/stretchr/testify/assert"
	"github.com/vinllen/mgo/bson"
)

const (
	TestMongoAddress          = "mongodb://100.81.164.177:30442,100.81.164.177:30441,100.81.164.177:30443"
	TestCheckpointDb          = "test_checkpoint_db"
	TestCheckpointTable       = "test_checkpoint_table"
	TestCheckpointStatusTable = "test_checkpoint_status_table"
)

func TestCheckpointCRUD(t *testing.T) {
	// test InsertCkpt, UpdateCkpt, QueryCkpt, DropCheckpoint
	ckptConn, err := utils.NewMongoConn(TestMongoAddress, utils.ConnectModePrimary, true)
	assert.Equal(t, nil, err, "should be equal")

	var nr int
	{
		fmt.Printf("TestCheckpointCRUD case %d.\n", nr)
		nr++

		// remove test checkpoint table
		err := DropCheckpoint(TestMongoAddress, TestCheckpointDb)
		assert.Equal(t, nil, err, "should be equal")

		_, err = QueryCkpt("test_id", ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, utils.NotFountErr, err.Error(), "should be equal")

		ckpt := &Checkpoint{
			ShardId:  "test_id",
			FatherId: "test_father",
			Status:   StatusNotProcess,
		}
		err = UpdateCkpt("test_id", ckpt, ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, utils.NotFountErr, err.Error(), "should be equal")

		err = UpdateCkptSet("test_id", bson.M{"status": StatusNotProcess}, ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, utils.NotFountErr, err.Error(), "should be equal")

		err = InsertCkpt(ckpt, ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")

		retCkpt, err := QueryCkpt("test_id", ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")
		assert.Equal(t, StatusNotProcess, retCkpt.Status, "should be equal")

		err = UpdateCkptSet("test_id", bson.M{"status": StatusInProcessing}, ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")

		retCkpt, err = QueryCkpt("test_id", ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")
		assert.Equal(t, StatusInProcessing, retCkpt.Status, "should be equal")
	}

	{
		fmt.Printf("TestCheckpointCRUD case %d.\n", nr)
		nr++

		// remove test checkpoint table
		err := DropCheckpoint(TestMongoAddress, TestCheckpointDb)
		assert.Equal(t, nil, err, "should be equal")

		ckpt := &Checkpoint{
			ShardId: "test_id_2",
			Status:  StatusPrepareProcess,
		}
		err = InsertCkpt(ckpt, ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")

		err = UpdateCkptSet("test_id_2", bson.M{"status": StatusDone}, ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")

		retCkpt, err := QueryCkpt("test_id_2", ckptConn, TestCheckpointDb, TestCheckpointTable)
		assert.Equal(t, nil, err, "should be equal")
		assert.Equal(t, StatusDone, retCkpt.Status, "should be equal")
	}
}

func TestExtractCheckpoint(t *testing.T) {
	// test ExtractCheckpoint, DropCheckpoint, InsertCkpt
	ckptConn, err := utils.NewMongoConn(TestMongoAddress, utils.ConnectModePrimary, true)
	assert.Equal(t, nil, err, "should be equal")

	var nr int
	{
		fmt.Printf("TestExtractCheckpoint case %d.\n", nr)
		nr++

		// remove test checkpoint table
		err := DropCheckpoint(TestMongoAddress, TestCheckpointDb)
		assert.Equal(t, nil, err, "should be equal")

		err = InsertCkpt(&Checkpoint{
			ShardId: "id1",
			Status:  StatusNotProcess,
		}, ckptConn, TestCheckpointDb, "table1")
		assert.Equal(t, nil, err, "should be equal")

		err = InsertCkpt(&Checkpoint{
			ShardId: "id2",
			Status:  StatusInProcessing,
		}, ckptConn, TestCheckpointDb, "table1")
		assert.Equal(t, nil, err, "should be equal")

		err = InsertCkpt(&Checkpoint{
			ShardId: "id3",
			Status:  StatusPrepareProcess,
		}, ckptConn, TestCheckpointDb, "table1")
		assert.Equal(t, nil, err, "should be equal")

		err = InsertCkpt(&Checkpoint{
			ShardId: "id1",
			Status:  StatusDone,
		}, ckptConn, TestCheckpointDb, "table2")
		assert.Equal(t, nil, err, "should be equal")

		err = InsertCkpt(&Checkpoint{
			ShardId: "id10",
			Status:  StatusWaitFather,
		}, ckptConn, TestCheckpointDb, "table2")
		assert.Equal(t, nil, err, "should be equal")

		mp, err := ExtractCheckpoint(ckptConn, TestCheckpointDb)
		assert.Equal(t, nil, err, "should be equal")
		assert.Equal(t, 2, len(mp), "should be equal")
		assert.Equal(t, 3, len(mp["table1"]), "should be equal")
		assert.Equal(t, 2, len(mp["table2"]), "should be equal")
		assert.Equal(t, StatusNotProcess, mp["table1"]["id1"].Status, "should be equal")
		assert.Equal(t, StatusInProcessing, mp["table1"]["id2"].Status, "should be equal")
		assert.Equal(t, StatusPrepareProcess, mp["table1"]["id3"].Status, "should be equal")
		assert.Equal(t, StatusDone, mp["table2"]["id1"].Status, "should be equal")
		assert.Equal(t, StatusWaitFather, mp["table2"]["id10"].Status, "should be equal")
	}

}

func TestCheckpointStatus(t *testing.T) {
	// test FindStatus, UpsertStatus, DropCheckpoint
	ckptConn, err := utils.NewMongoConn(TestMongoAddress, utils.ConnectModePrimary, true)
	assert.Equal(t, nil, err, "should be equal")

	var nr int
	{
		fmt.Printf("TestCheckpointStatus case %d.\n", nr)
		nr++

		// remove test checkpoint table
		err := DropCheckpoint(TestMongoAddress, TestCheckpointDb)
		assert.Equal(t, nil, err, "should be equal")

		status, err := FindStatus(ckptConn, TestCheckpointDb, TestCheckpointStatusTable)
		assert.Equal(t, nil, err, "should be equal")
		assert.Equal(t, CheckpointStatusValueEmpty, status, "should be equal")
	}

	{
		fmt.Printf("TestCheckpointStatus case %d.\n", nr)
		nr++

		// remove test checkpoint table
		err := DropCheckpoint(TestMongoAddress, TestCheckpointDb)
		assert.Equal(t, nil, err, "should be equal")

		err = UpsertStatus(ckptConn, TestCheckpointDb, TestCheckpointStatusTable, CheckpointStatusValueIncrSync)
		assert.Equal(t, nil, err, "should be equal")

		status, err := FindStatus(ckptConn, TestCheckpointDb, TestCheckpointStatusTable)
		assert.Equal(t, nil, err, "should be equal")
		assert.Equal(t, CheckpointStatusValueIncrSync, status, "should be equal")
	}
}
