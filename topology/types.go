// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package topology

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/m3db/m3cluster/client"
	"github.com/m3db/m3cluster/services"
	"github.com/m3db/m3db/sharding"
	"github.com/m3db/m3db/ts"
	"github.com/m3db/m3x/instrument"
)

// Host is a container of a host in a topology
type Host interface {
	// ID is the identifier of the host
	ID() string

	// Address returns the address of the host
	Address() string

	// String returns a string representation of the host
	String() string
}

// HostShardSet is a container for a host and corresponding shard set
type HostShardSet interface {
	// Host returns the host
	Host() Host

	// ShardSet returns the shard set owned by the host
	ShardSet() sharding.ShardSet
}

// Initializer can init new instances of Topology
type Initializer interface {
	// Init will return a new topology
	Init() (Topology, error)
}

// Topology is a container of a topology map and disseminates topology map changes
type Topology interface {
	// Get the topology map
	Get() Map

	// Watch for the topology map
	Watch() (MapWatch, error)

	// Close will close the topology map
	Close()
}

// DynamicTopology is a topology that dynamically changes and as such
// adds functionality for a clustered database to call back and mark
// a shard as available once it completes bootstrapping
type DynamicTopology interface {
	Topology

	// MarkShardAvailable marks a shard with the state of initializing as available
	MarkShardAvailable(instanceID string, shardID uint32) error
}

// MapWatch is a watch on a topology map
type MapWatch interface {
	// C is the notification channel for when a value becomes available
	C() <-chan struct{}

	// Get the current topology map
	Get() Map

	// Close the watch on the topology map
	Close()
}

// Map describes a topology
type Map interface {
	// Hosts returns all hosts in the map
	Hosts() []Host

	// HostShardSets returns all HostShardSets in the map
	HostShardSets() []HostShardSet

	// LookupHostShardSet returns a HostShardSet for a host in the map
	LookupHostShardSet(hostID string) (HostShardSet, bool)

	// HostsLen returns the length of all hosts in the map
	HostsLen() int

	// ShardSet returns the shard set for the topology
	ShardSet() sharding.ShardSet

	// Route will route a given ID to a shard and a set of hosts
	Route(id ts.ID) (uint32, []Host, error)

	// RouteForEach will route a given ID to a shard then execute a
	// function for each host in the set of routed hosts
	RouteForEach(id ts.ID, forEachFn RouteForEachFn) error

	// RouteShard will route a given shard to a set of hosts
	RouteShard(shard uint32) ([]Host, error)

	// RouteShardForEach will route a given shard and execute
	// a function for each host in the set of routed hosts
	RouteShardForEach(shard uint32, forEachFn RouteForEachFn) error

	// Replicas returns the number of replicas in the topology
	Replicas() int

	// MajorityReplicas returns the number of replicas to establish majority in the topology
	MajorityReplicas() int
}

// RouteForEachFn is a function to execute for each routed to host
type RouteForEachFn func(idx int, host Host)

// ConsistencyLevel is the consistency level for cluster operations
type ConsistencyLevel int

// nolint: deadcode, varcheck, unused
const (
	consistencyLevelNone ConsistencyLevel = iota

	// ConsistencyLevelOne corresponds to a single node participating
	// for an operation to succeed
	ConsistencyLevelOne

	// ConsistencyLevelMajority corresponds to the majority of nodes participating
	// for an operation to succeed
	ConsistencyLevelMajority

	// ConsistencyLevelAll corresponds to all nodes participating
	// for an operation to succeed
	ConsistencyLevelAll
)

// String returns the consistency level as a string
func (l ConsistencyLevel) String() string {
	switch l {
	case consistencyLevelNone:
		return "none"
	case ConsistencyLevelOne:
		return "one"
	case ConsistencyLevelMajority:
		return "majority"
	case ConsistencyLevelAll:
		return "all"
	}
	return "unknown"
}

var validConsistencyLevels = []ConsistencyLevel{
	consistencyLevelNone,
	ConsistencyLevelOne,
	ConsistencyLevelMajority,
	ConsistencyLevelAll,
}

var errConsistencyLevelUnspecified = errors.New("consistency level not specified")

// UnmarshalYAML unmarshals an ConnectConsistencyLevel into a valid type from string.
func (l *ConsistencyLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	if str == "" {
		return errConsistencyLevelUnspecified
	}
	strs := make([]string, len(validConsistencyLevels))
	for _, valid := range validConsistencyLevels {
		if str == valid.String() {
			*l = valid
			return nil
		}
		strs = append(strs, "'"+valid.String()+"'")
	}
	return fmt.Errorf("invalid ConsistencyLevel '%s' valid types are: %s",
		str, strings.Join(strs, ", "))
}

// StaticOptions is a set of options for static topology
type StaticOptions interface {
	// Validate validates the options
	Validate() error

	// SetShardSet sets the ShardSet
	SetShardSet(value sharding.ShardSet) StaticOptions

	// ShardSet returns the ShardSet
	ShardSet() sharding.ShardSet

	// SetReplicas sets the replicas
	SetReplicas(value int) StaticOptions

	// Replicas returns the replicas
	Replicas() int

	// SetHostShardSets sets the hostShardSets
	SetHostShardSets(value []HostShardSet) StaticOptions

	// HostShardSets returns the hostShardSets
	HostShardSets() []HostShardSet
}

// DynamicOptions is a set of options for dynamic topology
type DynamicOptions interface {
	// Validate validates the options
	Validate() error

	// SetConfigServiceClient sets the client of ConfigService
	SetConfigServiceClient(c client.Client) DynamicOptions

	// ConfigServiceClient returns the client of ConfigService
	ConfigServiceClient() client.Client

	// SetServiceID sets the ServiceID for service discovery
	SetServiceID(s services.ServiceID) DynamicOptions

	// ServiceID returns the ServiceID for service discovery
	ServiceID() services.ServiceID

	// SetServiceOptions sets the options for service discovery.
	SetServiceOptions(opts services.Options) DynamicOptions

	// ServicesOptions returns the options for service discovery.
	ServicesOptions() services.Options

	// SetQueryOptions sets the ConfigService query options
	SetQueryOptions(value services.QueryOptions) DynamicOptions

	// QueryOptions returns the ConfigService query options
	QueryOptions() services.QueryOptions

	// SetInstrumentOptions sets the instrumentation options
	SetInstrumentOptions(value instrument.Options) DynamicOptions

	// InstrumentOptions returns the instrumentation options
	InstrumentOptions() instrument.Options

	// SetInitTimeout sets the waiting time for dynamic topology to be initialized
	SetInitTimeout(value time.Duration) DynamicOptions

	// InitTimeout returns the waiting time for dynamic topology to be initialized
	InitTimeout() time.Duration

	// SetHashGen sets the HashGen function
	SetHashGen(h sharding.HashGen) DynamicOptions

	// HashGen returns HashGen function
	HashGen() sharding.HashGen
}