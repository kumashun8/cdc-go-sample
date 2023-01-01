# setupの実行ログ

## logical replication slotの作成

```sql
postgres=# SELECT pg_create_logical_replication_slot('replication_slot', 'test_decoding');
ERROR:  replication slot "replication_slot" already exists
postgres=# SELECT slot_name, plugin, slot_type, database, active, restart_lsn, confirmed_flush_lsn FROM pg_replication_slots;
    slot_name     |    plugin     | slot_type | database | active | restart_lsn | confirmed_flush_lsn 
------------------+---------------+-----------+----------+--------+-------------+---------------------
 replication_slot | test_decoding | logical   | postgres | f      | 0/153D818   | 0/153D850
(1 row)
```

## 確認用tableの作成

```sql

postgres=# create table kumatable1(
postgres(# kuma_id serial PRIMARY KEY,
postgres(# kuma_name VARCHAR(255) NOT NULL,
postgres(# role  VARCHAR(255) NOT NULL
postgres(# );
CREATE TABLE
postgres=# create table inutable1(
inu_id serial PRIMARY KEY,
inu_name VARCHAR(255) NOT NULL,
role VARCHAR(255) NOT NULL
);
CREATE TABLE
postgres=# SELECT * FROM pg_publication_tables WHERE pubname='pub';
 pubname | schemaname | tablename  |         attnames         | rowfilter 
---------+------------+------------+--------------------------+-----------
 pub     | public     | kumatable1 | {kuma_id,kuma_name,role} | 
 pub     | public     | inutable1  | {inu_id,inu_name,role}   | 
(2 rows)

```

## logical replicationの確認

```sql
postgres=# INSERT INTO kumatable1 (kuma_id, kuma_name, role)
postgres-# VALUES(1, 'shirokuma', 'student');
INSERT 0 1
postgres=# select * from kumatable1;
 kuma_id | kuma_name |  role   
---------+-----------+---------
       1 | shirokuma | student
(1 row)

postgres=# SELECT * FROM pg_logical_slot_get_changes('replication_slot', NULL, NULL);
    lsn    | xid |                                                              data                       
-----------+-----+-----------------------------------------------------------------------------------------
 0/1543650 | 736 | BEGIN 736
 0/1543930 | 736 | COMMIT 736
 0/1543A80 | 737 | BEGIN 737
 0/156BBA8 | 737 | COMMIT 737
 0/156BCF8 | 738 | BEGIN 738
 0/15905D8 | 738 | COMMIT 738
 0/15906F8 | 739 | BEGIN 739
 0/15906F8 | 739 | table public.kumatable1: INSERT: kuma_id[integer]:1 kuma_name[character varying]:'shirok
 0/1590818 | 739 | COMMIT 739
(9 rows)

postgres=# 
```
