	"github.com/m3db/m3db/persist/fs/msgpack"
	// ReadBloomFilter returns the bloom filter stored on disk in a container object that is safe
	// for concurrent use and has a Close() method for releasing resources when done.
	ReadBloomFilter() (*ManagedConcurrentBloomFilter, error)

	// SeekByID returns the data for specified ID provided the index was loaded upon open. An
	SeekByID(id ts.ID) (data checked.Bytes, err error)

	// SeekByIndexEntry is similar to Seek, but uses an IndexEntry instead of
	// looking it up on its own. Useful in cases where you've already obtained an
	// entry and don't want to waste resources looking it up again.
	SeekByIndexEntry(entry IndexEntry) (checked.Bytes, error)
	// SeekIndexEntry returns the IndexEntry for the specified ID. This can be useful
	// ahead of issuing a number of seek requests so that the seek requests can be
	// made in order. The returned IndexEntry can also be passed to SeekUsingIndexEntry
	// to prevent duplicate index lookups.
	SeekIndexEntry(id ts.ID) (IndexEntry, error)
	// ConcurrentIDBloomFilter returns a concurrency-safe bloom filter that can
	// be used to quickly disqualify ID's that definitely do not exist. I.E if the
	// Test() method returns true, the ID may exist on disk, but if it returns
	// false, it definitely does not.
	ConcurrentIDBloomFilter() *ManagedConcurrentBloomFilter
	// Validate will validate the options and return an error if not valid
	Validate() error

	// SetIndexSummariesPercent size sets the percent of index summaries to write
	SetIndexSummariesPercent(value float64) Options

	// IndexSummariesPercent size returns the percent of index summaries to write
	IndexSummariesPercent() float64

	// SetIndexBloomFilterFalsePositivePercent size sets the percent of false positive
	// rate to use for the index bloom filter size and k hashes estimation
	SetIndexBloomFilterFalsePositivePercent(value float64) Options

	// IndexBloomFilterFalsePositivePercent size returns the percent of false positive
	// rate to use for the index bloom filter size and k hashes estimation
	IndexBloomFilterFalsePositivePercent() float64


	// SetMmapEnableHugeTLB sets whether mmap huge pages are enabled when running on linux
	SetMmapEnableHugeTLB(value bool) Options

	// MmapEnableHugeTLB returns whether mmap huge pages are enabled when running on linux
	MmapEnableHugeTLB() bool

	// SetMmapHugeTLBThreshold sets the threshold when to use mmap huge pages for mmap'd files on linux
	SetMmapHugeTLBThreshold(value int64) Options

	// MmapHugeTLBThreshold returns the threshold when to use mmap huge pages for mmap'd files on linux
	MmapHugeTLBThreshold() int64