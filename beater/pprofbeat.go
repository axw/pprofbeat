package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/pkg/errors"

	"github.com/axw/pprofbeat/config"
	"github.com/axw/pprofbeat/internal/fetch"
)

type ProfileBeater struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	if c.URL == "" {
		return nil, errors.Errorf("%s.url must be specified", b.Info.Beat)
	}
	return &ProfileBeater{
		done:   make(chan struct{}),
		config: c,
	}, nil
}

func (pb *ProfileBeater) Stop() {
	pb.client.Close()
	close(pb.done)
}

func (pb *ProfileBeater) Run(b *beat.Beat) error {
	logp.Info("pprofbeat is running! Hit CTRL-C to stop it.")

	var err error
	pb.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(pb.config.Period)
	for {
		if err := pb.tick(b); err != nil {
			return err
		}
		select {
		case <-pb.done:
			return nil
		case <-ticker.C:
		}
	}
}

func (pb *ProfileBeater) tick(b *beat.Beat) error {
	// TODO(axw) should we use p.TimeNanos for the timestamp?
	timestamp := time.Now()
	logp.Info("fetching profile from %s", pb.config.URL)
	p, err := fetch.Fetch(fetch.Options{
		URL:      pb.config.URL,
		Duration: pb.config.FetchDuration,
		Timeout:  pb.config.FetchTimeout,
	})
	if err != nil {
		logp.Err("failed to fetch profile: %v", err)
		return nil
	}

	// NOTE(axw) samples contains highly redundant data for the
	// locations. Need to see how much storage costs balloon, or
	// how well the data is compressed and common data merged.
	fields := common.MapStr{"type": b.Info.Beat}
	if len(p.Comments) != 0 {
		fields["comments"] = p.Comments
	}
	if p.DurationNanos != 0 {
		fields["duration"] = time.Duration(p.DurationNanos).Seconds()
	}
	if p.Period != 0 {
		periodType := p.PeriodType.Type
		periodUnit := normalizeUnit(p.PeriodType.Unit)
		key := fmt.Sprintf("period.%s.%s", periodType, periodUnit)
		fields[key] = p.Period
	}
	if len(p.Sample) != 0 {
		functions := make(map[uint64]common.MapStr)
		for _, f := range p.Function {
			fmap := common.MapStr{
				"name": f.Name,
				"file": f.Filename,
			}
			if f.SystemName != f.Name {
				fmap["system_name"] = f.SystemName
			}
			if f.StartLine != 0 {
				fmap["start_line"] = f.StartLine
			}
			functions[f.ID] = fmap
		}

		mappings := make(map[uint64]common.MapStr)
		for _, m := range p.Mapping {
			mmap := common.MapStr{
				"start":  m.Start,
				"limit":  m.Limit,
				"offset": m.Offset,
				"file":   m.File,
			}
			if m.BuildID != "" {
				mmap["build_id"] = m.BuildID
			}
			if m.HasFunctions {
				mmap["has_functions"] = true
			}
			if m.HasFilenames {
				mmap["has_filenames"] = true
			}
			if m.HasLineNumbers {
				mmap["has_line_numbers"] = true
			}
			if m.HasInlineFrames {
				mmap["has_inline_frames"] = true
			}
			mappings[m.ID] = mmap
		}

		locations := make(map[uint64]common.MapStr)
		for _, l := range p.Location {
			lines := make([]common.MapStr, len(l.Line))
			for i, l := range l.Line {
				lines[i] = common.MapStr{
					"function": functions[l.Function.ID],
					"line":     l.Line,
				}
			}
			lmap := common.MapStr{
				"addr":  l.Address,
				"lines": lines,
			}
			if l.IsFolded {
				lmap["folded"] = true
			}
			locations[l.ID] = lmap
		}

		samples := make([]common.MapStr, len(p.Sample))
		for i, s := range p.Sample {
			sampleLocations := make([]common.MapStr, len(s.Location))
			for i, l := range s.Location {
				sampleLocations[i] = locations[l.ID]
			}
			sample := common.MapStr{
				"locations": sampleLocations,
				// TODO(axw) labels
			}
			for i, v := range s.Value {
				sampleType := p.SampleType[i].Type
				sampleUnit := normalizeUnit(p.SampleType[i].Unit)
				key := sampleType + "." + sampleUnit
				if key == "samples.count" {
					key = "count"
				}
				sample[key] = v
			}
			samples[i] = sample
		}
		fields["samples"] = samples
	}

	pb.client.Publish(beat.Event{
		Timestamp: timestamp,
		Fields:    fields,
	})
	logp.Info("Event sent")
	return nil
}

func normalizeUnit(unit string) string {
	switch unit {
	case "nanoseconds":
		unit = "ns"
	}
	return unit
}
