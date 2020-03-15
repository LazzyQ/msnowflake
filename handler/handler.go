package handler

import (
	"context"
	"github.com/LazzyQ/msnowflake/model/worker"
	msnowflake "github.com/LazzyQ/msnowflake/proto"
)


var (
	idWorder *worker.IdWorker
)

type MSnowflake struct {
	
}

func (m MSnowflake) NextId(ctx context.Context, req *msnowflake.IdRequest, res *msnowflake.IdResponse) error {
	id, err := idWorder.NextId()
	if err != nil {
		return err
	}
	res.Code = 0;
	res.Message = "success"
	res.Id = id
	return nil
}

func (m MSnowflake) NextIds(ctx context.Context, req *msnowflake.IdRequest, res *msnowflake.IdResponse) error {
	ids, err := idWorder.NextIds(req.Num)
	if err != nil {
		return err
	}
	res.Code = 0;
	res.Message = "success"
	res.Ids = ids
	return nil
}

func Init() (err error)  {
	idWorder, err = worker.GetIdWorker()
	return err
}

