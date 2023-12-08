package post

import (
	"reflect"
	"testing"

	"github.com/greatjudge/redditclone/pkg/comment"
	"github.com/greatjudge/redditclone/pkg/user"
	"github.com/greatjudge/redditclone/pkg/vote"
	"github.com/stretchr/testify/assert"
)

func CheckInit(t *testing.T, pType string) {
	usr := user.User{
		ID:       "1",
		Username: "username",
	}
	post := Post{}
	post.Type = pType
	if pType == TEXT {
		post.URL = "sdksld"
	} else {
		post.Text = "kdslfkdlf"
	}
	InitPost(&post, usr)

	assert.Equal(t, post.Author, usr)
	if pType == TEXT {
		assert.Equal(t, post.URL, "")
	} else {
		assert.Equal(t, post.Text, "")
	}
	assert.Equal(t, post.Votes, []vote.Vote{
		{
			UserID: usr.ID,
			Value:  1,
		},
	})
	assert.Equal(t, post.Score, 1)
	assert.Equal(t, post.UpvotePercentage, 100)
}

func TestInitPost(t *testing.T) {
	CheckInit(t, TEXT)
	CheckInit(t, "url")
}

type TestCase struct {
	expectedScore            int
	expectedUpvotePercentage int
	expectedVotes            []vote.Vote
	p                        Post
	method                   string
	idArg                    string
	casename                 string
}

func Check(t *testing.T, tc TestCase) {
	p := tc.p
	switch tc.method {
	case "upvote":
		p.Upvote(tc.idArg)
	case "downvote":
		p.Downvote(tc.idArg)
	case "unvote":
		p.Unvote(tc.idArg)
	}

	if p.Score != tc.expectedScore {
		t.Errorf("bad Score. expected: %d, got %d", tc.expectedScore, p.Score)
	}
	if p.UpvotePercentage != tc.expectedUpvotePercentage {
		t.Errorf("bad UpvotePercentage. expected %d, got %d", tc.expectedUpvotePercentage, p.UpvotePercentage)
	}
	if !reflect.DeepEqual(p.Votes, tc.expectedVotes) {
		t.Errorf("bad votes. Expected %v,\n got %v", tc.expectedVotes, tc.expectedUpvotePercentage)
	}
}

func TestUpvote(t *testing.T) {
	p := Post{
		ID:    "1",
		Title: "title",
		Views: 2,
		Type:  "text",
		Author: user.User{
			ID:       "1",
			Username: "username",
		},
		Category: "category",
		Votes: []vote.Vote{
			{
				UserID: "1",
				Value:  1,
			},
		},
		Comments:         make([]comment.Comment, 0),
		Created:          CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	}

	cases := []TestCase{
		{
			p:                        p,
			expectedScore:            p.Score + 1,
			expectedUpvotePercentage: 100,
			expectedVotes: []vote.Vote{
				{
					UserID: "1",
					Value:  1,
				},
				{
					UserID: "2",
					Value:  1,
				},
			},
			idArg:    "2",
			method:   "upvote",
			casename: "new vote",
		},
		{
			p:                        p,
			expectedScore:            p.Score,
			expectedUpvotePercentage: 100,
			expectedVotes: []vote.Vote{
				{
					UserID: "1",
					Value:  1,
				},
			},
			idArg:    "1",
			method:   "upvote",
			casename: "exist vote",
		},
	}

	p.Score = -1
	p.UpvotePercentage = 0
	p.Votes = []vote.Vote{
		{
			UserID: "1",
			Value:  -1,
		},
	}

	cases = append(cases, TestCase{
		p:                        p,
		expectedScore:            1,
		expectedUpvotePercentage: 100,
		expectedVotes: []vote.Vote{
			{
				UserID: "1",
				Value:  1,
			},
		},
		idArg:    "1",
		method:   "upvote",
		casename: "opposite vote",
	})

	for _, tc := range cases {
		t.Run(tc.casename, func(t *testing.T) {
			Check(t, tc)
		})
	}
}

func TestDownvote(t *testing.T) {
	p := Post{
		ID:    "1",
		Title: "title",
		Views: 2,
		Type:  "text",
		Author: user.User{
			ID:       "1",
			Username: "username",
		},
		Category: "category",
		Votes: []vote.Vote{
			{
				UserID: "1",
				Value:  1,
			},
		},
		Comments:         make([]comment.Comment, 0),
		Created:          CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	}

	cases := []TestCase{
		{
			p:                        p,
			expectedScore:            p.Score - 1,
			expectedUpvotePercentage: 50,
			expectedVotes: []vote.Vote{
				{
					UserID: "1",
					Value:  1,
				},
				{
					UserID: "2",
					Value:  -1,
				},
			},
			idArg:    "2",
			method:   "downvote",
			casename: "new vote",
		},
		{
			p:                        p,
			expectedScore:            -1,
			expectedUpvotePercentage: 0,
			expectedVotes: []vote.Vote{
				{
					UserID: "1",
					Value:  -1,
				},
			},
			idArg:    "1",
			method:   "downvote",
			casename: "opposite vote",
		},
	}

	p.Score = -1
	p.UpvotePercentage = 0
	p.Votes = []vote.Vote{
		{
			UserID: "1",
			Value:  -1,
		},
	}

	cases = append(cases, TestCase{
		p:                        p,
		expectedScore:            -1,
		expectedUpvotePercentage: 0,
		expectedVotes: []vote.Vote{
			{
				UserID: "1",
				Value:  -1,
			},
		},
		idArg:    "1",
		method:   "downvote",
		casename: "exist vote",
	})

	for _, tc := range cases {
		t.Run(tc.casename, func(t *testing.T) {
			Check(t, tc)
		})
	}
}

func TestUnvote(t *testing.T) {
	p := Post{
		ID:    "1",
		Title: "title",
		Views: 2,
		Type:  "text",
		Author: user.User{
			ID:       "1",
			Username: "username",
		},
		Category: "category",
		Votes: []vote.Vote{
			{
				UserID: "1",
				Value:  1,
			},
		},
		Comments:         make([]comment.Comment, 0),
		Created:          CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	}

	cases := []TestCase{
		{
			p:                        p,
			expectedScore:            p.Score - 1,
			expectedUpvotePercentage: 0,
			expectedVotes:            []vote.Vote{},
			idArg:                    "1",
			method:                   "unvote",
			casename:                 "exist +1 vote",
		},
		{
			p:                        p,
			expectedScore:            p.Score,
			expectedUpvotePercentage: 100,
			expectedVotes: []vote.Vote{
				{
					UserID: "1",
					Value:  1,
				},
			},
			idArg:    "2",
			method:   "unvote",
			casename: "no vote",
		},
	}

	p.Score = -1
	p.UpvotePercentage = 0
	p.Votes = []vote.Vote{
		{
			UserID: "1",
			Value:  -1,
		},
	}

	cases = append(cases, TestCase{
		p:                        p,
		expectedScore:            p.Score + 1,
		expectedUpvotePercentage: 0,
		expectedVotes:            []vote.Vote{},
		idArg:                    "1",
		method:                   "unvote",
		casename:                 "exist -1 vote",
	})

	for _, tc := range cases {
		t.Run(tc.casename, func(t *testing.T) {
			Check(t, tc)
		})
	}
}
