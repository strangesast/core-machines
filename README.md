# Aggregations

## Recent Activity
```
select
	*,
	(case when next_timestamp is null then now() else next_timestamp end) - timestamp as diff
from (
	select
		*,
		lead(timestamp, 1) over (partition by machine_id order by timestamp asc) as next_timestamp,
		dense_rank() over (partition by machine_id order by timestamp asc) as r
	from (
		select machine_id, property, value, timestamp::timestamp from machine_state
	) t
	where property = 'execution'
) t
where value = 'ACTIVE'
order by timestamp desc
```

## Part Counts
```
select *
from (
	select
		*,
		dense_rank() over (partition by machine_id, value order by timestamp asc) as r
	from (
		select 
			machine_id,
			property,
			timestamp::timestamp,
			value
		from machine_state
	) t
	where property = 'part_count' and value != 'UNAVAILABLE'
) t
where r = 1
order by timestamp desc
```

## Execution Periods
```
select
	*,
	(case when next_timestamp is null then now() else next_timestamp end) - timestamp
from (
	select
		*,
		lead(timestamp, 1) over (partition by machine_id order by timestamp asc) as next_timestamp
	from (
		select
			machine_id,
			property,
			value,
			to_timestamp(timestamp/1000) as timestamp
		from machine_state
	) t
	where property = 'execution'
) t
order by timestamp desc
```

## Active Execution Periods
```
select
	machine_id,
	property,
	timestamp,
	value,
	case when next_timestamp is null then now() else next_timestamp end,
	r
from (
	select
		*,
		dense_rank() over (partition by machine_id order by timestamp asc) as r,
		lead(timestamp, 1) over (partition by machine_id order by timestamp asc) as next_timestamp
	from (
		select
			machine_id,
			property,
			timestamp::timestamp,
			value 
		from machine_state
		where property = 'execution'
	) t
) t
where value = 'ACTIVE'
order by timestamp desc
```

```
select
	a.*,
	b.r
from (
	select
		machine_id,
		property,
		timestamp::timestamp,
		value
	from machine_state
	where property = 'part_count'
) a
join (
	select
		machine_id,
		timestamp,
		case when next_timestamp is null then now() else next_timestamp end,
		r
	from (
		select
			*,
			dense_rank() over (partition by machine_id order by timestamp asc) as r,
			lead(timestamp, 1) over (partition by machine_id order by timestamp asc) as next_timestamp
		from (
			select
				machine_id,
				property,
				timestamp::timestamp,
				value 
			from machine_state
			where property = 'execution'
		) t
	) t
	where value = 'ACTIVE'
) b on (
	a.machine_id = b.machine_id and
	a.timestamp >= b.timestamp and
	a.timestamp < b.next_timestamp
)
order by a.timestamp desc
```

# Machine Values During ACTIVE Executions
```
select *
from (
	select
		machine_id,
		property,
		to_timestamp(timestamp/1000) as timestamp,
		value,
		"offset"
	from machine_values
) a
join (
	select *
	from machine_execution_state
	where value = 'ACTIVE'
	order by timestamp desc
	limit 10
) b
on (a.machine_id = b.machine_id and a.timestamp >= b.timestamp and a.timestamp < b.next_timestamp)
```

## Part Count Active Execution
```
select *
from (
	select
		machine_id,
		value,
		timestamp,
		case when next_timestamp is null then now() else next_timestamp end,
		"offset"
	from machine_execution_state
	where value = 'ACTIVE'
	order by timestamp desc
) a
left join machine_part_count b on (
	a.machine_id = b.machine_id and b.timestamp >= a.timestamp and b.timestamp < a.next_timestamp
)
```
