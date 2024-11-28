## sflight

Another singleflight implementation with Generics and expires(in case you forgot to call ```group.Forget```).


## Installation

```bash
go get github.com/snownd/sflight
```

## Usage

```go
g := sflight.New[int, string](5 * time.Second)
v, err, isExecuted := g.Do(1, func() (string, error) {
    return "v", nil
})
if err!= nil {
    return err
}

// optional Forget
g.Forget(1)
```


