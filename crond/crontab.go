package crond

type Crontab struct {
	entries []Entry
	name    string
}

func NewCrontab(name string, entries []Entry) *Crontab {
	return &Crontab{
		name:    name,
		entries: entries,
	}
}

func (c *Crontab) Generate() error {
	return nil
}
