# CURL Wrapper

## Usage

```
curl2 -json -bearer key1 [other curl args] https://example.com
```

## Flags

- `-json`
- `-bearer`
- `-basic`

## Flag Usage

### -json

This will add the `Content-Type` and `Accept` header

### -bearer

Appends the Bearer Token Authentication header. Used in conjunction with `kv`. The _value_ of the corresponding key will
be used to construct the authorization header.

```
curl2 -bearer key1 https://example.com
```

### -basic

Same as `-bearer` for Basic Authentication

