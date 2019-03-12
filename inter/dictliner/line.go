package dictliner

import (
	"github.com/peterh/liner"
)

// Line 封装 *liner.State, 提供更简便的操作
type DictLiner struct {
	State   *liner.State
	History *DictLineHistory

	tmode liner.ModeApplier
	lmode liner.ModeApplier

	paused bool
}

// NewLine 返回 *DictLiner, 默认设置允许 Ctrl+C 结束
func NewLiner() *DictLiner {
	pl := &DictLiner{}
	pl.tmode, _ = liner.TerminalMode()

	line := liner.NewLiner()
	pl.lmode, _ = liner.TerminalMode()

	line.SetMultiLineMode(true)
	line.SetCtrlCAborts(true)

	pl.State = line

	return pl
}

// Pause 暂停服务
func (pl *DictLiner) Pause() error {
	if pl.paused {
		panic("DictLiner already paused")
	}

	pl.paused = true
	pl.DoWriteHistory()

	return pl.tmode.ApplyMode()
}

// Resume 恢复服务
func (pl *DictLiner) Resume() error {
	if !pl.paused {
		panic("DictLiner is not paused")
	}

	pl.paused = false

	return pl.lmode.ApplyMode()
}

// Close 关闭服务
func (pl *DictLiner) Close() (err error) {
	err = pl.State.Close()
	if err != nil {
		return err
	}

	if pl.History != nil && pl.History.historyFile != nil {
		return pl.History.historyFile.Close()
	}

	return nil
}
