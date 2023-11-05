# Configuring TTL: Guidelines and Instructions

- **TTL_INTERVAL**: This parameter sets the numeric value for the TTL duration. For instance, if you wish for data to expire after a duration of 2 time units, set this value to 2. The default value is 1.

- **TTL_UNIT**: This parameter specifies the time unit for the TTL duration. It accepts valid values such as SECOND, MINUTE, HOUR, DAY, MONTH, and more. For example, to set a TTL of 2 hours, you would set TTL_INTERVAL to 2 and TTL_UNIT to HOUR. The default unit is MONTH.

# Usage

## Setting Environment Variables

To configure TTL for your application, set the desired environment variables. Here's an example of how to do this:

```bash
export TTL_INTERVAL=5
export TTL_UNIT=MINUTE
```

