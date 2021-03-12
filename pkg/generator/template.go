package generator

import "text/template"

var tpl = template.Must(template.New("generated").
	Funcs(template.FuncMap{
		"lcFirst": lcFirst,
	}).
	Parse(`
// Code generated by github.com/shenjing023/dataloaden, DO NOT EDIT.

package {{.Package}}

import (
    "sync"
    "time"
	"context"

    {{if .KeyType.ImportPath}}"{{.KeyType.ImportPath}}"{{end}}
    {{if .ValType.ImportPath}}"{{.ValType.ImportPath}}"{{end}}
)

// {{.Name}}Config captures the config to create a new {{.Name}}
type {{.Name}}Config struct {
	// Fetch is a method that provides the data for the loader 
	Fetch func(ctx context.Context, keys []{{.KeyType.String}}) ([]{{.ValType.String}}, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

type cacheItem struct {
	expire time.Time
	value  {{.ValType.String}}
}

// New{{.Name}} creates a new {{.Name}} given a fetch, wait, and maxBatch
func New{{.Name}}(config {{.Name}}Config) *{{.Name}} {
	return &{{.Name}}{
		fetch: config.Fetch,
		wait: config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// {{.Name}} batches and caches requests          
type {{.Name}} struct {
	// this method provides the data for the loader
	fetch func(ctx context.Context, keys []{{.KeyType.String}}) ([]{{.ValType.String}}, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[{{.KeyType.String}}]*cacheItem

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *{{.Name|lcFirst}}Batch

	// mutex to prevent races
	mu sync.Mutex

	// expiration time
	expiration time.Duration
}

type {{.Name|lcFirst}}Batch struct {
	keys    []{{.KeyType}}
	data    []{{.ValType.String}}
	error   []error
	closing bool
	done    chan struct{}
}

// Load a {{.ValType.Name}} by key, batching and caching will be applied automatically
func (l *{{.Name}}) Load(ctx context.Context, key {{.KeyType.String}}) ({{.ValType.String}}, error) {
	return l.LoadThunk(ctx,key)()
}

// LoadThunk returns a function that when called will block waiting for a {{.ValType.Name}}.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *{{.Name}}) LoadThunk(ctx context.Context, key {{.KeyType.String}}) func() ({{.ValType.String}}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if it, ok := l.cache[key]; ok {
		if l.expiration > 0 {
			// had set expire
			if time.Now().Before(it.expire) {
				// not expire
				return func() ({{.ValType.String}}, error) {
					return it.value, nil
				}
			}
			delete(l.cache, key)
		} else {
			return func() ({{.ValType.String}}, error) {
				return it.value, nil
			}
		}
	}

	if l.batch == nil {
		l.batch = &{{.Name|lcFirst}}Batch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(ctx, l, key)
	l.mu.Unlock()

	return func() ({{.ValType.String}}, error) {
		<-batch.done

		var data {{.ValType.String}}
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *{{.Name}}) LoadAll(ctx context.Context, keys []{{.KeyType}}) ([]{{.ValType.String}}, []error) {
	results := make([]func() ({{.ValType.String}}, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(ctx,key)
	}

	{{.ValType.Name|lcFirst}}s := make([]{{.ValType.String}}, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		{{.ValType.Name|lcFirst}}s[i], errors[i] = thunk()
	}
	return {{.ValType.Name|lcFirst}}s, errors
}

// LoadAllThunk returns a function that when called will block waiting for a {{.ValType.Name}}s.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *{{.Name}}) LoadAllThunk(ctx context.Context, keys []{{.KeyType}}) (func() ([]{{.ValType.String}}, []error)) {
	results := make([]func() ({{.ValType.String}}, error), len(keys))
 	for i, key := range keys {
		results[i] = l.LoadThunk(ctx,key)
	}
	return func() ([]{{.ValType.String}}, []error) {
		{{.ValType.Name|lcFirst}}s := make([]{{.ValType.String}}, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			{{.ValType.Name|lcFirst}}s[i], errors[i] = thunk()
		}
		return {{.ValType.Name|lcFirst}}s, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *{{.Name}}) Prime(key {{.KeyType}}, value {{.ValType.String}}) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		{{- if .ValType.IsPtr }}
			// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
			// and end up with the whole cache pointing to the same value.
			cpy := *value
			l.unsafeSet(key, &cpy)
		{{- else if .ValType.IsSlice }}
			// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
			// and end up with the whole cache pointing to the same value.
			cpy := make({{.ValType.String}}, len(value))
			copy(cpy, value)
			l.unsafeSet(key, cpy)
		{{- else }}
			l.unsafeSet(key, value)
		{{- end }}
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *{{.Name}}) Clear(key {{.KeyType}}) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *{{.Name}}) unsafeSet(key {{.KeyType}}, value {{.ValType.String}}) {
	if l.cache == nil {
		l.cache = map[{{.KeyType}}]*cacheItem{}
	}
	l.cache[key] = &cacheItem{
		expire: time.Now().Add(l.expiration),
		value:  value,
	}
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *{{.Name|lcFirst}}Batch) keyIndex(ctx context.Context, l *{{.Name}}, key {{.KeyType}}) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(ctx, l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(ctx,l)
		}
	}

	return pos
}

func (b *{{.Name|lcFirst}}Batch) startTimer(ctx context.Context, l *{{.Name}}) {
	// wait all goroutine append key to b.keys
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(ctx,l)
}

func (b *{{.Name|lcFirst}}Batch) end(ctx context.Context, l *{{.Name}}) {
	b.data, b.error = l.fetch(ctx,b.keys)
	close(b.done)
}
`))
