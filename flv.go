package record

import (
	"path/filepath"
	"strconv"
	"time"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
)

type FLVRecorder struct {
	Recorder
}

func (r *FLVRecorder) Start(streamPath string) (err error) {
	r.Record = &RecordPluginConfig.Flv
	r.ID = streamPath + "/flv"
	if _, ok := RecordPluginConfig.recordings.Load(r.ID); ok {
		return ErrRecordExist
	}
	return plugin.Subscribe(streamPath, r)
}

func (r *FLVRecorder) start() {
	RecordPluginConfig.recordings.Store(r.ID, r)
	r.PlayFLV()
	RecordPluginConfig.recordings.Delete(r.ID)
	r.Close()
}

func (r *FLVRecorder) OnEvent(event any) {
	switch v := event.(type) {
	case ISubscriber:
		filename := strconv.FormatInt(time.Now().Unix(), 10) + r.Ext
		if r.Fragment == 0 {
			filename = r.Stream.Path + r.Ext
		} else {
			filename = filepath.Join(r.Stream.Path, filename)
		}
		if file, err := r.CreateFileFn(filename, r.append); err == nil {
			r.SetIO(file)
		}
		// 写入文件头
		if !r.append {
			r.Write(codec.FLVHeader)
		}
		go r.start()
	case FLVFrame:
		if r.Fragment > 0 {
			check := false
			if r.Video.Track == nil {
				check = true
			} else {
				check = r.Video.Frame.IFrame
			}
			if ts := r.Video.Frame.AbsTime - r.SkipTS; check && int64(ts) >= int64(r.Fragment*1000) {
				r.SkipTS = r.Video.Frame.AbsTime
				r.Close()
				if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
					r.SetIO(file)
					r.Write(codec.FLVHeader)
					if r.Video.Track != nil {
						dcflv := codec.VideoAVCC2FLV(r.Video.Track.DecoderConfiguration.AVCC, 0)
						dcflv.WriteTo(r)
					}
					if r.Audio.Track != nil && r.Audio.Track.CodecID == codec.CodecID_AAC {
						dcflv := codec.AudioAVCC2FLV(r.Audio.Track.Value.AVCC, 0)
						dcflv.WriteTo(r)
					}
					flv := codec.VideoAVCC2FLV(r.Video.Frame.AVCC, 0)
					flv.WriteTo(r)
					return
				}
			}
		}

		if _, err := v.WriteTo(r); err != nil {
			r.Stop()
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
