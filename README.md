# Deathquake Go

A real-time Quake 3 Arena game statistics monitor written in Go. Tracks player kills, deaths, rankings, and scores during live gameplay.

## Requirements

- **Go 1.22.2** or higher
- **ioquake3** server installation
- **gawk** for log timestamping

## Quick Start

### 1. Start the Quake 3 Server

**Important:** Before starting your event, edit `start_server.sh` and change the rcon password from the default `Hunter2` to a secure password.

```bash
./start_server.sh /path/to/ioquake3
```

This will:
- Copy `server.cfg` to the ioquake3 directory and execute it when the server starts
- Set the rcon password in `start_server.sh` for remote administration (see [ioquake3 rcon documentation](https://ioquake3.org/help/sys-admin-guide/#rcon) for executing commands from the game client)
- Create a timestamped log file in the current directory: `game_YYYYMMDD_HHMMSS.log`

### 2. Run Deathquake Go

```bash
go run main.go -f <game_YYYYMMDD_HHMMSS.log>
```

Or build and run:
```bash
go build -o deathquake
./deathquake -f game_20251206_143022.log
```

### Debug Mode

View detailed logging output:
```bash
./deathquake -f game.log --debug
```

## Game Rules

For an example of how to structure game rules for Deathquake events, see [SAMPLE_RULES.md](SAMPLE_RULES.md). This document contains sample scoring and drinking game rules that can be adapted for your own events.

## Event Winners

See [WINNERS.md](WINNERS.md) for a list of past winners from Deathquake events held at the Department of Computer Science, Aarhus University.

## Configuration

Edit `config.json` to customize behavior:

```json
{
  "ignored_players": ["PlayerName"],
  "drinking_cider_players": ["PlayerName"],
  "ignored_rounds": []
}
```

- **ignored_players**: Players to exclude from statistics (Note: `<world>` is always ignored automatically)
- **drinking_cider_players**: Players using special scoring mode
- **ignored_rounds**: Round hashes to ignore (found in debug mode output)

### Ignoring Rounds

You can configure Deathquake Go to ignore specific round sessions by their hash.

#### Finding Round Hashes

1. Run the application in debug mode:
   ```bash
   ./deathquake -f game.log --debug
   ```

2. The round hash appears in all log lines after a map change:
   ```
   [5d41402abc4b2a76b9719d911017c592] [MAP] Round ID generated: 5d41402abc4b2a76b9719d911017c592
   [5d41402abc4b2a76b9719d911017c592] [KILL] PlayerOne killed PlayerTwo with MOD_RAILGUN
   [5d41402abc4b2a76b9719d911017c592] [SAVE] Saving round results
   [5d41402abc4b2a76b9719d911017c592] [SAVE] Frag limit for this round: 20
   ```

3. Copy the hash from the square brackets at the beginning of any line

#### Adding Hashes to Ignore Rounds

Add the round hash to `config.json`:

```json
{
  "ignored_rounds": [
    "5d41402abc4b2a76b9719d911017c592",
    "7d793037a0760186574b0282f2f435e7"
  ]
}
```

Once added, rounds with matching hashes will be ignored during parsing.

#### Warmup Behavior
- **The first map is always treated as warmup** - statistics are not recorded until a map change occurs
- **Live tracking begins after the first map change** - once the second map loads, the game becomes active
- **A game session ends when the scoreboard appears** - triggered by reaching the time limit or frag limit configured in server.cfg

If multiple scoreboards appear on the same map, only the first one counts as a game end. The map must change to exit warmup mode and begin a new session.

## Known Limitations

### Player Names Containing "killed"

If a player's name contains the string "killed", kill events involving that player will be ignored. This is due to the log file parsing logic, which searches for the "killed" keyword to identify kill events. Multiple occurrences of "killed" in a single line cause parsing ambiguity and the line will be skipped.

**Example of problematic names:**
- `Player killed`
- `killed_by_noob`
- `killedYou123`

## Advanced Usage

### Combining Log Files for Streaming

If you have multiple log files and want to replay them as a continuous stream (e.g., for testing or replay), you can combine them using `cat` and `tail`:

```bash
# Create a streaming log from two files
(cat game_20251206_143022.log && tail -f game_20251206_150000.log) > combined_stream.log
```

Then in another terminal:
```bash
./deathquake -f combined_stream.log
```

**How it works:**
1. `cat game_20251206_143022.log` - Outputs the entire first log file immediately
2. `&&` - Waits for the first command to complete, then runs the next
3. `tail -f game_20251206_150000.log` - Follows the second log file in real-time
4. `> combined_stream.log` - Redirects all output to a new combined file

**Use cases:**
- Continue monitoring after a server crash
- Maintain continuous statistics across server restarts
- Avoid losing game session data when the dedicated server restarts

## License

This project is licensed under the BEERWARE License (Revision 42).

## Contributing

Contributions are welcome! Please ensure all tests pass before submitting pull requests.
