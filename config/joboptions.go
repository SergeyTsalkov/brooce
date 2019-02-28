package config

import (
	"log"
	"time"

	"brooce/mergo"
	"github.com/mitchellh/copystructure"
)

// in the JSON, setting an option to false/0 is different than omitting it
// only omitted values should be inherited, so we have to use pointers to tell
// the difference
type JobOptions struct {
	Timeout_        *int  `json:"timeout,omitempty"`
	MaxTries_       *int  `json:"maxtries,omitempty"`
	KillOnDelay_    *bool `json:"killondelay,omitempty"`
	NoFail_         *bool `json:"nofail,omitempty"`
	RequeueDelayed_ *int  `json:"requeuedelayed,omitempty"`
	RequeueFailed_  *int  `json:"requeuefailed,omitempty"`
	//PruneDone_   *int  `json:"prunedone,omitempty"`

	NoRedisLog_                 *bool `json:"noredislog,omitempty"`
	NoRedisLogOnSuccess_        *bool `json:"noredislogonsuccess,omitempty"`
	NoRedisLogOnFail_           *bool `json:"noredislogonfail,omitempty"`
	RedisLogExpireAfter_        *int  `json:"redislogexpireafter,omitempty"`
	RedisLogFailedExpireAfter_  *int  `json:"redislogfailedexpireafter,omitempty"`
	RedisListDoneExpireAfter_   *int  `json:"redislistdoneexpireafter,omitempty"`
	RedisListFailedExpireAfter_ *int  `json:"redislistfailedexpireafter,omitempty"`

	Drop_          *bool `json:"drop,omitempty"`
	DropOnSuccess_ *bool `json:"droponsuccess,omitempty"`
	DropOnFail_    *bool `json:"droponfail,omitempty"`
}

var DefaultJobOptions = JobOptions{}

func initDefaultJobOptions() {
	DefaultJobOptions.Timeout_ = intptr(3600)
	DefaultJobOptions.MaxTries_ = intptr(1)
	DefaultJobOptions.RedisLogExpireAfter_ = intptr(604800) // 7 days
	DefaultJobOptions.RequeueDelayed_ = intptr(60)
}

func intptr(x int) *int {
	return &x
}

func (j *JobOptions) Timeout() int {
	if j.Timeout_ != nil && *j.Timeout_ > 0 {
		return *j.Timeout_
	}
	return 3600
}

func (j *JobOptions) TimeoutDuration() time.Duration {
	return time.Duration(j.Timeout()) * time.Second
}

func (j *JobOptions) MaxTries() int {
	if j.MaxTries_ != nil && *j.MaxTries_ > 0 {
		return *j.MaxTries_
	}
	return 1
}

func (j *JobOptions) KillOnDelay() bool {
	return j.KillOnDelay_ != nil && *j.KillOnDelay_
}

func (j *JobOptions) NoFail() bool {
	return j.NoFail_ != nil && *j.NoFail_
}

func (j *JobOptions) RequeueDelayed() int {
	if j.RequeueDelayed_ != nil && *j.RequeueDelayed_ > 0 {
		return *j.RequeueDelayed_
	}
	return 60
}

func (j *JobOptions) RequeueFailed() int {
	if j.RequeueFailed_ != nil && *j.RequeueFailed_ > 0 {
		return *j.RequeueFailed_
	}
	return 0
}

/*
func (j *JobOptions) PruneDone() int {
	if j.PruneDone_ != nil && *j.PruneDone_ > 0 {
		return *j.PruneDone_
	}
	return 0
}
*/

func (j *JobOptions) NoRedisLog() bool {
	if j.NoRedisLogOnSuccess() && j.NoRedisLogOnFail() {
		return true
	}
	return j.NoRedisLog_ != nil && *j.NoRedisLog_
}

func (j *JobOptions) NoRedisLogOnSuccess() bool {
	if j.NoRedisLog_ != nil && *j.NoRedisLog_ {
		return true
	}

	return j.NoRedisLogOnSuccess_ != nil && *j.NoRedisLogOnSuccess_
}

func (j *JobOptions) NoRedisLogOnFail() bool {
	if j.NoRedisLog_ != nil && *j.NoRedisLog_ {
		return true
	}

	return j.NoRedisLogOnFail_ != nil && *j.NoRedisLogOnFail_
}

func (j *JobOptions) RedisLogExpireAfter() int {
	if j.RedisLogExpireAfter_ != nil && *j.RedisLogExpireAfter_ > 0 {
		return *j.RedisLogExpireAfter_
	}
	return 0
}

func (j *JobOptions) RedisLogFailedExpireAfter() int {
	if j.RedisLogFailedExpireAfter_ != nil && *j.RedisLogFailedExpireAfter_ > 0 {
		return *j.RedisLogFailedExpireAfter_
	}
	return 0
}

func (j *JobOptions) RedisListDoneExpireAfter() int {
	if j.RedisListDoneExpireAfter_ != nil && *j.RedisListDoneExpireAfter_ > 0 {
		return *j.RedisListDoneExpireAfter_
	}
	return 0
}

func (j *JobOptions) RedisListFailedExpireAfter() int {
	if j.RedisListFailedExpireAfter_ != nil && *j.RedisListFailedExpireAfter_ > 0 {
		return *j.RedisListFailedExpireAfter_
	}
	return 0
}

func (j *JobOptions) Drop() bool {
	if j.DropOnSuccess() && j.DropOnFail() {
		return true
	}

	return j.Drop_ != nil && *j.Drop_
}

func (j *JobOptions) DropOnSuccess() bool {
	if j.Drop_ != nil && *j.Drop_ {
		return true
	}

	return j.DropOnSuccess_ != nil && *j.DropOnSuccess_
}

func (j *JobOptions) DropOnFail() bool {
	if j.Drop_ != nil && *j.Drop_ {
		return true
	}

	return j.DropOnFail_ != nil && *j.DropOnFail_
}

func (j *JobOptions) clone() JobOptions {
	jci, err := copystructure.Copy(*j)
	if err != nil {
		log.Fatalln("Copy structure error:", err)
	}

	return jci.(JobOptions)
}

func (j *JobOptions) Merge(parent JobOptions) {
	// don't want to copy pointers to values in parent -- we might change those values later, which would
	// inadvertently change parent
	if err := mergo.Merge(j, parent.clone()); err != nil {
		log.Fatalf("merge wtf: %+v", err)
	}
}
