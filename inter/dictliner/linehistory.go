package dictliner

import (
	"fmt"
	"os"
)

// LineHistory 命令行历史
type DictLineHistory struct {
	historyFilePath string
	historyFile     *os.File
}

// NewLineHistory 设置历史
func NewLineHistory(filePath string) (lh *DictLineHistory, err error) {
	lh = &DictLineHistory{
		historyFilePath: filePath,
	}

	lh.historyFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return lh, nil
}

// DoWriteHistory 执行写入历史
func (pl *DictLiner) DoWriteHistory() (err error) {
	if pl.History == nil {
		return fmt.Errorf("history not set")
	}

	pl.History.historyFile, err = os.Create(pl.History.historyFilePath)
	if err != nil {
		return fmt.Errorf("写入历史错误, %s", err)
	}

	_, err = pl.State.WriteHistory(pl.History.historyFile)
	if err != nil {
		return fmt.Errorf("写入历史错误: %s", err)
	}

	return nil
}

// ReadHistory 读取历史
func (pl *DictLiner) ReadHistory() (err error) {
	if pl.History == nil {
		return fmt.Errorf("history not set")
	}

	_, err = pl.State.ReadHistory(pl.History.historyFile)
	return err
}
