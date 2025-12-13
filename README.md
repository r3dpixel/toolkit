# Toolkit

A Go utility library with common packages for everyday programming tasks.

## Packages

### async

Non-blocking context cancellation checks. Lets you check if a context is cancelled without blocking the current
goroutine.

### bytex

Human-readable byte sizes (KB, MB, GB) with parsing and formatting. Also includes reusable 32KB buffer pools to reduce
memory allocations.

### cred

Credential storage using either OS keyring or environment variables. Store and retrieve usernames/passwords securely.

### filex

File operations - check if files/dirs exist, copy files efficiently, sanitize filenames by removing invalid characters.

### imagex

Load and save images in any format. Handles encoding/decoding between PNG, JPEG, GIF, etc.

### jsonx

JSON operations with generics. Marshal/unmarshal to/from files and bytes with type safety. Uses high-performance Sonic
library.

### ptr

Create pointers from values easily. Just convenience functions for working with pointers.

### reqx

HTTP client with automatic retries, JWT refresh, and browser impersonation. Handles auth tokens that expire and need
refreshing.

### scheduler

Parallel task execution with two modes: pull-based (Exec) for processing tasks from a source, and push-based (Pool) for
submitting tasks at runtime. Both support configurable parallelism and context cancellation.

### slicesx

Generic slice operations - map, merge, grow. Functional programming patterns for slices.

### sonicx

Wrapper around Sonic JSON parser. Navigate JSON using paths like "user.name.first" and extract values.

### stringsx

String utilities - check if blank, join non-empty strings, normalize weird Unicode quotes to ASCII.

### symbols

Constants for common symbols and pre-compiled regex patterns used across the toolkit.

### structx

Just provides an empty struct for use in map[string]struct{} when you need a set.

### templater

Simple templating engine that replaces tokens like {{name}} with values. Uses a trie for fast token matching.

### timestamp

Work with Unix timestamps at different precisions (seconds, milliseconds, microseconds, nanoseconds). Parse and format
time strings.

### trace

Errors with metadata fields and chaining. Create errors that carry context like user_id, request_id, etc.

### util

Random utilities that don't fit elsewhere. Currently just has GetOrDefault.