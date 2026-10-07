# Goblin - a Go chaos proxy - WIP

The purpose of `Goblin` is to create chaos with systems over HTTP protocols.  The intent is to test with specific end points and not be used as a broader tool hence there is not route specifc options.  It is all or none which aligns with how I'm currently validating my custom applications.  Routing *may* be a future consideration along with other protocols aside from just HTTP.

## Running

### Initialize
```bash
./goblin -init
```
that will create your first configuration with default values

### Run
```bash
./goblin
```

### Admin panel
`http://localhost:8080/gablin`

Fair warning, it is ugly.

From the admin panel you can update the existing settings.  Providing you enter in valid settings the application will automatically update.  There is also a flag that can be set to overwrite the existing `goblin.toml` file so that your settings are saved for the next time you start the application.

The interface is basic as it is only used to adjust settings.

## Building

simple build command
```bash
make build
```

## Testing

The test application are dumb on purpose and are only meant to validate configuration

1. Start the application
```bash
make run
```
2. start the reciever
```bash
make listen
```

3. start the sender
```bash
make sender
```

## Future
- add more policies as needed/desired
- add gRPC

### Contact
Greg D'Angelo
[LinkedIn](https://linkedin.com/in/gregdangelo)
