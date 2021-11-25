# REGEX Extraction

The `regex` command can be used to apply regex on each line of standard in and print the 
captured groups as CSV.

```
cat somefile.txt | regex regex1 [regex2] [regex3]
```

if there are multiple regex arguments, the first matching is picked. 