# Utilbox Transform

## 1. CSV

Used for transforming csv or screen scarped data.

eg. restart all kubernetes deployments

```
kubectl get deployment | csv col[0] row[1:] | \
 while read -r; do kubectl rollout restart deployment $REPLY; done
```

### Usage

```
cat pods.txt | csv [flag] [flag]
cat pods.txt | csv col[1]
```

### Flags

- `row`
- `col`
- `split`
- `merge`
- `tr`
- `group`
- `calc`
- `sort`
- `head`
- `-inhead`
- `-outhead`
- `out`
- `lmerge`

### Flag Description

#### row

The rows to select

```
row[1:]
row[1:5]
```

#### col

The columns to select

```
col[1:]
col[1:5]
col[1,2,3,4]
col[1-10,20-30]
```

The inverse option can be selected by `ncol`

#### split

The split delimiter while reading the data

```
split:<delim_str>
split:<special_str>
```

The `special_str` options are

```
csv     => use the csv reader to read data
comma   => ,
space   => 
tab     => \t
newline => \n
quote   => '
dquote  => "
none    =>
pipe    => |
```

#### merge

The merge delimiter for the data output

```
merge:<delim_str>
merge:<special_str>
```

The `special_str` options are same as `split`

#### tr

Transforms the columns

```
tr..c5..split:/..merge:-..col[-1]..pfx:^..sfx:$..add

tr..c5..split:/..merge:-..col[0]  tr..c5..split::..merge:-..col[0]  => chained transformation 
```

The same column transformed again by chaining, the second `tr` will be applied on the output of first `tr`
unless `add` flag is set. If the add flag is `set` then the `tr` will append a new col and leave the existing column
intact

where,

- `..`            tr arg delimiter; any same set of chars succeeding the `tr` will be used as delim eg `tr#col[1]`
  or `tr:::col[1]`
- `c<col_num>`    `col_num` is the column on which the transform is applied
- `split`         Split Delimiter
- `merge`         Merge Delimiter
- `col` or `ncol` The col indices to pick or exclude
- `pfx`           Add any prefix string
- `sfx`           Add any suffix strings
- `add`           Apply transform and create a new col. _default_ is to modify the same column

#### group

Group the data based on the column indices. The column indices used are based on the output. This is applied after the
column transformations. As a part of grouping the number data will be added and string will be concatenated.

```
group[0,1,2]
group[1]:count   => count will add a count column
```

#### calc

Apply some formulae on the rows. The column indices used are based on the output, not input. This will attempt to
convert the str data into number. This is applied after `group` operation

```
calc([0]+[1])
'calc([0]+"/"+[1])'
```

#### sort

Sorts the data. Will attempt to convert the str data into number. Sorting is a final operation, so the column indices
are based on the output data, not the input

```
sort[2,1]
sort[2,1]:asc
sort[2,1]:desc
```

#### head

Provide a set of new column headers

```
head[col1,col2,col3]
```

In case of **JSON** output, additional header names can be set to normalize the output into a tree structure based on
the _grouping_. It will infer the _level count_ based on the count of additional headers.

#### -inhead
_minus inhead_

This indicates that there is no header in the input. The headers for the output can be added using the `head[...]` flag 

#### -outhead
_minus outhead_

Omit headers while printing the output. Valid only for `table` and `csv` output formats 

#### out

The output format

```
output..csv     => Default
output..json    
output..table
```

- `..` is the arg delimiter same case as `tr`

Note: **JSON output** has some additional options

```
out..json..levels:<level_count>
```

where `level_count` is the depth value to _normalize_ the JSON output into a tree structure based on _grouping_. The max
value of the levels should the be count of `group_by` columns.

#### lmerge

Merge all the lines into one line in the output

```
lmerge
lmerge:<delim_str>
```

The `special_str` options are same as `split`

## 2 JSON

Transforms JSON input. The command is `jp` i.e _JSONProcessor_

### Usage

```
# PRINTS THE JSON KEYS
kubectl get pods -o json | jp keys

# GENERATE DATA BASED ON THE SELECTED KEYS 
kubectl get pods -o json | jp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP]
```

### Flags

- `keys`
- `out`
- `head`
- `-outhead`
- `sort`
- `calc`

### Flag Usage

Refer to CSV section for common flag usage

#### keys

```
jp keys            => List the keys

jp keys[key1,key2] => selects the values based on keys into a table 
```
Note: For _grouping_ pipe the `jp` output into `csv` and transform it further. 

## 3 YAML

Transforms Yaml input. The command is `yp` i.e _YamlProcessor_. Flags and behavior are identical to the JSON variant

### Usage

```
# PRINTS THE YAML KEYS
kubectl get pods -o yaml | yp keys

# GENERATE DATA BASED ON THE SELECTED KEYS 
kubectl get pods -o yaml | yp keys[items.metadata.name,items.metadata.namespace,items.status.hostIP,items.status.podIP]
```

### Flags

- `keys`
- `out`
- `head`
- `-outhead`
- `sort`
- `calc`

### Flag Usage

Refer to CSV section for common flag usage

#### keys

```
yp keys            => List the keys

yp keys[key1,key2] => selects the values based on keys into a table 
```
Note: For _grouping_ pipe the `yp` output into `csv` and transform it further. 