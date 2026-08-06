package common

type FeedURL string

func (u FeedURL) String() string {
	return string(u)
}
