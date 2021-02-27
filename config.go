package mongo

// NOTE: read config when call Init(env)

var DevCnf = map[string][]string{
	"gdc": []string{
		"username:password@dev-ip.:27017",
	},
}

var TestCnf = map[string][]string{
	"gdc": []string{
		"username:password@test-ip.:27017",
	},
}

var ProdCnf = map[string][]string{
	"gdc": []string{
		"username:password@prod-ip.:27017",
	},
}
