package record

import (
	"bufio"
	"io"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
)

type IRecorder interface {
	ISubscriber
	GetRecorder() *Recorder
	Start(streamPath string) error
	StartWithFileName(streamPath string, fileName string) error
	io.Closer
	CreateFile() (FileWr, error)
	StartWithDynamicTimeout(streamPath, fileName string, timeout time.Duration) error
	UpdateTimeout(timeout time.Duration)
}

type Recorder struct {
	Subscriber
	SkipTS   uint32
	Record   `json:"-" yaml:"-"`
	File     FileWr `json:"-" yaml:"-"`
	FileName string // 自定义文件名，分段录像无效
	filePath string // 文件路径
	append   bool   // 是否追加模式
}

func (r *Recorder) GetRecorder() *Recorder {
	return r
}

func (r *Recorder) CreateFile() (f FileWr, err error) {
	r.filePath = r.getFileName(r.Stream.Path) + r.Ext
	f, err = r.CreateFileFn(r.filePath, r.append)
	logFields := []zap.Field{zap.String("path", r.filePath)}
	if fw, ok := f.(*FileWriter); ok && r.Config != nil {
		if r.Config.WriteBufferSize > 0 {
			logFields = append(logFields, zap.Int("bufferSize", r.Config.WriteBufferSize))
			fw.bufw = bufio.NewWriterSize(fw.Writer, r.Config.WriteBufferSize)
			fw.Writer = fw.bufw
		}
	}
	if err == nil {
		r.Info("create file", logFields...)
	} else {
		logFields = append(logFields, zap.Error(err))
		r.Error("create file", logFields...)
	}
	return
}

func (r *Recorder) getFileName(streamPath string) (filename string) {
	filename = streamPath
	if r.Fragment == 0 {
		if r.FileName != "" {
			filename = filepath.Join(filename, r.FileName)
		}
	} else {
		filename = filepath.Join(filename, strings.ReplaceAll(streamPath, "/", "-")+"-"+time.Now().Format("2006-01-02-15-04-05"))
	}
	return
}

func (r *Recorder) start(re IRecorder, streamPath string, subType byte) (err error) {
	err = plugin.Subscribe(streamPath, re)
	if err == nil {
		if _, loaded := RecordPluginConfig.recordings.LoadOrStore(r.ID, re); loaded {
			return ErrRecordExist
		}
		r.Closer = re
		go func() {
			r.PlayBlock(subType)
			RecordPluginConfig.recordings.Delete(r.ID)
		}()
	}
	return
}

func (r *Recorder) cut(absTime uint32) {
	if ts := absTime - r.SkipTS; time.Duration(ts)*time.Millisecond >= r.Fragment {
		r.SkipTS = absTime
		r.Close()
		if file, err := r.Spesific.(IRecorder).CreateFile(); err == nil {
			r.File = file
			r.Spesific.OnEvent(file)
		} else {
			r.Stop(zap.Error(err))
		}
	}
}

// func (r *Recorder) Stop(reason ...zap.Field) {
// 	r.Close()
// 	r.Subscriber.Stop(reason...)
// }

func (r *Recorder) OnEvent(event any) {
	switch v := event.(type) {
	case IRecorder:
		if file, err := r.Spesific.(IRecorder).CreateFile(); err == nil {
			r.File = file
			r.Spesific.OnEvent(file)
		} else {
			r.Stop(zap.Error(err))
		}
	case AudioFrame:
		// 纯音频流的情况下需要切割文件
		if r.Fragment > 0 && r.VideoReader == nil {
			r.cut(v.AbsTime)
		}
	case VideoFrame:
		if v.IFrame {
			//go func() { //将视频关键帧的数据存入sqlite数据库中
			//	var flvKeyfram = &FLVKeyframe{FLVFileName: r.Path + "/" + strings.ReplaceAll(r.filePath, "\\", "/"), FrameOffset: r.VideoReader, FrameAbstime: v.AbsTime}
			//	sqlitedb.Create(flvKeyfram)
			//}()
			r.Info("这是关键帧，且取到了r.filePath是" + r.Path + r.filePath)
			//r.Info("这是关键帧，且取到了r.VideoReader.AbsTime是" + strconv.FormatUint(uint64(v.FrameAbstime), 10))
			//r.Info("这是关键帧，且取到了r.Offset是" + strconv.Itoa(int(v.FrameOffset)))
			//r.Info("这是关键帧，且取到了r.Offset是" + r.Stream.Path)
		}
		if r.Fragment > 0 && v.IFrame {
			r.cut(v.AbsTime)
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
