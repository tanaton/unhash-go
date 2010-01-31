package unhash2

/* ハッシュ値詰め合わせ */
type UnhashBox struct {
	hash	uint32;				/* ハッシュ値 */
	sum		uint32;				/* 確認用 */
}

/* 木構造 */
type UnhashTree struct {
	tree	[2]interface{};		/* 枝を二つまで管理 */
}

/* 連結リスト */
type UnhashData struct {
	box		UnhashBox;			/* ハッシュ値 */
	data	interface{};		/* 保存データ */
	next	*UnhashData;		/* 次のリスト */
}

/* unhash情報管理 */
type Unhash struct {
	tree		*UnhashTree;	/* 最初の枝 */
	max_level	uint8;			/* 最大階層 */
	table		[257]uint32;
}

const (
	UNHASH_PRIMES_TABLE = 257;
	UNHASH_HEAP_ARRAY_SIZE = 16;
	UNHASH_CACHE_SIZE = 0x0F;
	UNHASH_POLY = 0x9c55b4ee;
)

/* Unhashオブジェクト生成・初期化 */
func NewUnhash(ml uint8) (*Unhash){
	if ml > 32 || ml < 4 {
		/* 64bitより大きく3bit以下の場合 */
		return nil;
	}
	/* Unhashオブジェクト確保 */
	list := new(Unhash);
	/* tree領域確保 */
	list.tree = new(UnhashTree);
	/* 最大深度セット */
	list.max_level = ml;
	list.table = [UNHASH_PRIMES_TABLE]uint32{
		   2,    3,    5,    7,   11,   13,   17,   19,   23,   29,
		  31,   37,   41,   43,   47,   53,   59,   61,   67,   71,
		  73,   79,   83,   89,   97,  101,  103,  107,  109,  113,
		 127,  131,  137,  139,  149,  151,  157,  163,  167,  173,
		 179,  181,  191,  193,  197,  199,  211,  223,  227,  229,
		 233,  239,  241,  251,  257,  263,  269,  271,  277,  281,
		 283,  293,  307,  311,  313,  317,  331,  337,  347,  349,
		 353,  359,  367,  373,  379,  383,  389,  397,  401,  409,
		 419,  421,  431,  433,  439,  443,  449,  457,  461,  463,
		 467,  479,  487,  491,  499,  503,  509,  521,  523,  541,
		 547,  557,  563,  569,  571,  577,  587,  593,  599,  601,
		 607,  613,  617,  619,  631,  641,  643,  647,  653,  659,
		 661,  673,  677,  683,  691,  701,  709,  719,  727,  733,
		 739,  743,  751,  757,  761,  769,  773,  787,  797,  809,
		 811,  821,  823,  827,  829,  839,  853,  857,  859,  863,
		 877,  881,  883,  887,  907,  911,  919,  929,  937,  941,
		 947,  953,  967,  971,  977,  983,  991,  997, 1009, 1013,
		1019, 1021, 1031, 1033, 1039, 1049, 1051, 1061, 1063, 1069,
		1087, 1091, 1093, 1097, 1103, 1109, 1117, 1123, 1129, 1151,
		1153, 1163, 1171, 1181, 1187, 1193, 1201, 1213, 1217, 1223,
		1229, 1231, 1237, 1249, 1259, 1277, 1279, 1283, 1289, 1291,
		1297, 1301, 1303, 1307, 1319, 1321, 1327, 1361, 1367, 1373,
		1381, 1399, 1409, 1423, 1427, 1429, 1433, 1439, 1447, 1451,
		1453, 1459, 1471, 1481, 1483, 1487, 1489, 1493, 1499, 1511,
		1523, 1531, 1543, 1549, 1553, 1559, 1567, 1571, 1579, 1583,
		1597, 1601, 1607, 1609, 1613, 1619, 1621	};
	return list;
}

/* Unhashオブジェクトに値をセットする */
func (l *Unhash) Set(key string, data interface{}){
	box := UnhashBox{0, 0};
	/* hashを計算する */
	l.hashCreate(key, &box);
	tmp := l.areaGet(&box);
	if (tmp.box.hash != box.hash) || (tmp.box.sum != box.sum) {
		/* 連結リストをたどる */
		tmp = dataNext(tmp, &box);
	}
	tmp.data = data;
}

/* Unhashオブジェクトから値を取得 */
func (l *Unhash) Get(key string) (interface{}){
	box := UnhashBox{0, 0};
	/* hashを計算する */
	l.hashCreate(key, &box);
	/* Unhashツリーから探索 */
	data := l.areaGet(&box);
	if (data.box.hash != box.hash) || (data.box.sum != box.sum) {
		/* 連結リスト上を探索 */
		data = dataNext(data, &box);
	}
	return data.data;
}

/* Unhashツリーを生成、値の格納場所を用意 */
func (l *Unhash) areaGet(box *UnhashBox) (*UnhashData){
	var tree interface{} = l.tree;
	var data *UnhashData;
	var rl uint32 = 0;
	node := box.hash;
	for i := uint32(l.max_level - 1); i > 0; i-- {
		rl = (node >> i) & 0x01;	/* 方向選択 */
		if tree.(*UnhashTree).tree[rl] == nil {
			tree.(*UnhashTree).tree[rl] = new(UnhashTree);
		}
		tree = tree.(*UnhashTree).tree[rl];
	}
	/* 最深部 */
	rl = node & 0x01;
	if tree.(*UnhashTree).tree[rl] == nil {
		/* UnhashDataオブジェクト生成 */
		data = new(UnhashData);
		data.box = UnhashBox{box.hash, box.sum};
		tree.(*UnhashTree).tree[rl] = data;
	} else {
		data = tree.(*UnhashTree).tree[rl].(*UnhashData);
	}
	/* UnhashDataオブジェクトを返す */
	return data;
}

/* 連結リストをたどる */
func dataNext(data *UnhashData, box *UnhashBox) (*UnhashData){
	tmp := data.next;
	back := data;
	for tmp != nil {
		if((tmp.box.hash == box.hash) && (tmp.box.sum == box.sum)){
			/* それっぽいものを発見 */
			return tmp;
		} else {
			/* 次へ */
			back = tmp;
			tmp = tmp.next;
		}
	}
	/* UnhashDataオブジェクト生成 */
	tmp = new(UnhashData);
	tmp.box = UnhashBox{box.hash, box.sum};
	back.next = tmp;
	return tmp;
}

/* hash値生成 */
func (l *Unhash) hashCreate(str string, box *UnhashBox){
	length := len(str);
	var hash, sum, tmp uint32 = uint32(str[0]), 0, 0;
	for i := 0; i < length; i++ {
		tmp = uint32(str[i]);
		hash = hash * l.table[hash % UNHASH_PRIMES_TABLE] + tmp;
		sum = sum * 37 + tmp;
	}
	box.hash = hash;
	box.sum = sum;
}

