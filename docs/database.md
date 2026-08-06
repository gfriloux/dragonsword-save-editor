# Database schema

Once decrypted ([Encryption](encryption.md)), the save is an ordinary SQLite database
with **105 `tb_*` tables**. It holds a single account, so almost every table is keyed by
`USER_DBID` (the account id, e.g. `1000`). Timestamps are Unix seconds (`INTEGER`) or
`TEXT` datetimes depending on the table.

## ID conventions

Two kinds of ids recur throughout:

- **`*_CID` — content id.** A stable identifier that references the game's static data
  (an item, character, stat, region…). Small integers (5–7 digits). These are the ids
  documented in [Content IDs](content-ids.md) and used by th.gl.
- **`*_DBID` — instance id.** A unique, per-instance 64-bit id for a concrete owned
  object (a specific equipment piece, cooked dish, gem…). These **exceed 2^53**, so tools
  that store them as floating-point numbers (e.g. JavaScript) lose precision — handle
  them as 64-bit integers or strings.

## Gameplay tables (documented)

### `tb_user` — the account/player (1 row)

Player-wide state. Notable columns: `REGION_CID`, `SECTION_UID` (current area),
`POS_X/Y/Z` and `UNSAFE_POS_X/Y/Z` (world position), `ROTATION`, `CUR_STAMINA` /
`REWARD_MAX_STAMINA`, `LAST_LOGIN_TIME` / `LAST_LOGOUT_TIME` / `CREATE_DATE`,
`SELECTED_TEAM_PAGE_ID` / `SELECTED_PAGE_SLOT_INDEX`, `USER_DB_VERSION`.

### `tb_character` — owned characters

`(USER_DBID, CHARACTER_CID)` PK. `LEVEL`, `EXP`, `ASCEND`, `HP`, `TRANSCEND`,
`TRANSCEND_TOTAL`, `SOLDIER_GRADE`, `SOLDIER_GRADE_POINT`, `CREATED_DATE`.
`CHARACTER_CID` is a 5-digit CID (e.g. `10001` = Eileen).

### `tb_team` / `tb_init_team` — team pages

`(USER_DBID, PAGE_ID)` PK; `SLOT1/2/3_CHARACTER_CID` reference `tb_character`. A `0` slot
is empty. `tb_init_team` is the initial/default composition.

### `tb_currency` — currencies

`(USER_DBID, ITEM_CID)` PK, `AMOUNT`. Currency CIDs are in the `1000xxx` range.

### `tb_stackable_item` — stackable inventory

`(USER_DBID, ITEM_CID)` PK, `STACK_CNT`. Materials and stackable consumables (potions,
etc.). Keyed by `ITEM_CID`, so a `INSERT OR REPLACE` adds or sets a stack.

### `tb_cook_item` — cooked food (instances)

`(USER_DBID, ITEM_DBID)` PK. `ITEM_CID` = the dish, `SPECIAL_BUFF_CID1/2` = extra buffs,
`STACK_CNT`, `CREATED_DATE`, `DELETED_DATE` (0 = present). Each cooked stack is an
instance with its own 64-bit `ITEM_DBID`.

### `tb_equipment` — equipment (instances)

`ITEM_DBID` PK. `ITEM_CID` = the gear, `ENCHANT_LEVEL`, `EXP`, `IS_LOCK` (0/1),
`GEM_DBID` (socketed gem, → `tb_gem`), `MAIN_STAT_CID` and `SUB_STAT_CID1..5` (stat
references), `CREATED_DATE`, `DELETED_DATE` (0 = present, active rows only).

### `tb_gem` — gems (instances)

`ITEM_DBID` PK. `ITEM_CID`, `STAT_INFO_CID`, `IS_LOCK`, dates. Holds *socketed/instanced*
gems (often empty; unsocketed gems are stackable inventory items).

### `tb_costume` / `tb_vehicle` — cosmetics & mounts

**`tb_costume`** — one row per owned costume. `COSTUME_DBID` PK (instance id, fits a
uint32), `COSTUME_CID` (the cosmetic, `999xxxx`), `EQUIP_CHARACTER_CID` (the wearer's
character CID, `0` = not worn), `PARTS_ON` (visible-parts bitmask, not an on/off flag —
stays `1` even when equipped), `IS_NEW` (unread badge). `EQUIP_CHARACTER_CID` is `NOT
NULL` **without a default**, so an insert must supply it; `CREATED_DATE`/`PARTS_ON`/
`IS_NEW` default to now/1/1. Equip is **not exclusive**: several rows may point to the
same character at once — the `999xxxx` space bundles both outfits and weapon skins (see
[content-ids](content-ids.md)), and a character wears an outfit + its matching weapon
skin simultaneously.

**`tb_vehicle`** — one row per owned mount (familier). `VEHICLE_DBID` PK, `VEHICLE_CID`
(`132xxxx`), `CREATED_DATE`.

**`tb_equip_mount`** — `(USER_DBID, CHARACTER_CID)` PK, one row per character. `VEHICLE`
holds the equipped mount's `VEHICLE_DBID` (`0` = none); the `ACC_*`, `TALISMAN_*` and
`KARMA` columns hold that character's mount accessories. A character may have no row at
all, so equipping a mount upserts (insert with `0` defaults, or update `VEHICLE` in place
keeping the accessories).

The editor exposes both as **Costumes** and **Familiers** screens (unlock + equip);
`tb_summon` (an unrelated, counter-shaped summon-collection table) is out of scope.

### `tb_title` — account titles

`(USER_DBID, CATEGORY)` PK, plus `BIT_FIELD` and `FAV_BIT_FIELD` (both 64-bit masks). A
**bitmask** collection like `tb_switch`: whether a title is *unlocked* is one bit, with
`CATEGORY = title_id / 64` and `bit = title_id % 64`. The 108 titles come from
`AccountTitleData.xml` in the paks (ids `2100000`–`2104xxx`, 7 categories; see
[content-ids](content-ids.md)); each also carries a font colour and a small stat bonus.
`BIT_FIELD` holds the unlocked set; `FAV_BIT_FIELD` the displayed/favourite title.

The editor's **Titres** screen unlocks titles (per-title checkbox or unlock-all). Writes
go through an UPSERT that sets `BIT_FIELD` only, so `FAV_BIT_FIELD` is preserved — unlike
`tb_switch`, `tb_title` has that extra column, so a plain `INSERT OR REPLACE` would wipe
it. Editing the favourite/displayed title is out of scope. Category `32843` uses bits past
62, so masks are held as `uint64` and stored as `int64` (two's complement).

### `tb_quick_slot` — quick slots

`(USER_DBID, SLOT_INDEX)` PK, `ITEM_DBID`, `ITEM_CID`.

### `tb_skill_growth` — character skills

`(USER_DBID, CHARACTER_CID, TYPE_VALUE)` PK, `SLOT_LEVEL`.

### `tb_switch` and friends — bitmask flags

`(USER_DBID, CATEGORY)` PK, `BIT_FIELD` (a 64-bit bitmask). A generic flag system used
for many "collections", including **known cooking recipes**. See
[Switches & recipes](switches.md). Related tables (`tb_*_switch`,
`tb_karma_collection_switch`, `tb_episode_switch`…) follow the same shape for other
content types.

## Full table list (105)

Row counts below are from one real save, only to hint at what a table holds.

**Player / account** — `tb_user` (1), `tb_init_user`, `tb_play_point`, `tb_title` (6),
`tb_reward_step`, `tb_screenshot`, `tb_timer`, `tb_reset`, `tb_global_reset_values`.

**Characters & team** — `tb_character` (10), `tb_init_character` (23), `tb_team` (5),
`tb_init_team` (6), `tb_skill_growth` (49), `tb_summon`, `tb_equip_preset` (5).

**Inventory & items** — `tb_stackable_item` (111), `tb_cook_item` (23),
`tb_equipment` (56), `tb_gem`, `tb_costume` (10), `tb_vehicle` (14),
`tb_equip_mount` (8), `tb_quick_slot` (2), `tb_currency` (2), `tb_storage`,
`tb_treasure_box` (55), `tb_karma` (287), `tb_karma_collection_switch` (4).

**Switches / flags** — `tb_switch` (48, includes recipes) and the periodic/typed
variants: `tb_switch_day/week/month`, `tb_unexpected_switch_*`, `tb_reminiscence_switch`,
`tb_episode_switch`, `tb_event_rewarded_switch`, `tb_mercenary_trainee_*_switch`.

**Progression / quests** — `tb_quest_complete` (107), `tb_quest_hold` (3),
`tb_completed_quest`, `tb_quest_record_reward` (8), `tb_dynamic_quest_*`,
`tb_mission_event_progress`, `tb_multi_mission`, `tb_achievement_count` (58),
`tb_achievement_category_step` (3), `tb_trial_tower` (1), `tb_world_reputation`,
`tb_advent_status` (1), `tb_field_statue` (2), `tb_field_statue_charge` (1),
`tb_minigame` (8), `tb_mercenary_*`, `tb_vehicle_mission` (7), `tb_level_reach_event`,
`tb_attendance_event`, `tb_shop_event`.

**World / dungeons** — `tb_active_dungeon` (16), `tb_crack_dungeon*`,
`tb_user_instance_region` (18), `tb_actor_respawn`, `tb_spawn_condition_actor` (51),
`tb_map_memo`.

**Economy / shops** — `tb_npc_shop*`, `tb_bm_shop_cnt*` (black-market),
`tb_storage_bm_shop`, `tb_pass`, `tb_subscription`, `tb_purchase_play_point_by_cash_cnt`,
`tb_play_point`.

**Requests / social / misc** — `tb_acceptable_request`, `tb_account_buff`,
`tb_contents_buff*`, `tb_contents_playcount_*`, `tb_chat_*`, `tb_mail_state`,
`tb_expired_info`, `tb_client_binary`, `tb_test`, `tb_test_sync`.

> Tables not detailed above are present but **not yet documented** with certainty. This
> reference grows as we confirm more.
