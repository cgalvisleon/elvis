package linq

func (c *Linq) join(kind string, from, join *FRom, where *Where) *Linq {
	c.SetFromAs(from)
	c.SetFromAs(join)

	val1 := Col{
		from: join.as,
		name: c.Col(where.val1).name,
		cast: c.Col(where.val1).cast,
	}

	val2 := Col{
		from: from.as,
		name: c.Col(where.val2).name,
		cast: c.Col(where.val2).cast,
	}

	_where := &Where{
		connector: where.connector,
		val1:      val1,
		operator:  where.operator,
		val2:      val2,
	}

	result := &Join{
		kind:  kind,
		from:  from,
		join:  join,
		where: _where,
	}

	c._join = append(c._join, result)

	return c
}

func (c *Linq) Join(from, join *FRom, where *Where) *Linq {
	where.linq = c
	return c.join("INNER JOIN", from, join, where)
}

func (c *Linq) LeftJoin(from, join *FRom, where *Where) *Linq {
	where.linq = c
	return c.join("LEFT JOIN", from, join, where)
}

func (c *Linq) RightJoin(from, join *FRom, where *Where) *Linq {
	where.linq = c
	return c.join("RIGHT JOIN", from, join, where)
}
