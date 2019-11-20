package company

type Company struct {
	Name  string `json:"name"`
	Tel   string `json:"tel"`
	Email string `json:"email"`
}

var database = map[string]Company{"hung": Company{Name: "hung", Tel: "098", Email: "hung@silentium.io"}}

type CompanyRepository struct {
}

func NewCompanyRepository() *CompanyRepository {
	return &CompanyRepository{}
}

func (c *CompanyRepository) FindAll() []Company {
	items := make([]Company, 0, len(database))
	for _, v := range database {
		items = append(items, v)
	}
	return items
}

func (c *CompanyRepository) FindBy(key string) (Company, bool) {
	com, ok := database[key]
	return com, ok
}

func (c *CompanyRepository) Save(key string, item Company) {
	database[key] = item
}

func (c *CompanyRepository) Remove(key string) {
	delete(database, key)
}
