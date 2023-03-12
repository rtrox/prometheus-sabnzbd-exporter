package exporter

import (
	"prometheus-sabnzbd-exporter/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateServerStatCache_SameDay(t *testing.T) {
	require := require.New(t)
	cache := ServerStatCache{
		todayKey:                  "2020-01-01",
		total:                     1,
		articlesTriedHistorical:   2,
		articlesTriedToday:        3,
		articlesSuccessHistorical: 2,
		articlesSuccessToday:      3,
	}
	cache.Update(models.ServerStat{
		Total:           2,
		ArticlesTried:   5,
		ArticlesSuccess: 5,
		DayParsed:       "2020-01-01",
	})
	require.Equal(2, cache.GetTotal())
	require.Equal(7, cache.GetArticlesTried())
	require.Equal(7, cache.GetArticlesSuccess())
}

func TestUpdateServerStatCache_NewDay(t *testing.T) {
	require := require.New(t)
	cache := ServerStatCache{
		todayKey:                  "2020-01-01",
		total:                     1,
		articlesTriedHistorical:   2,
		articlesTriedToday:        3,
		articlesSuccessHistorical: 2,
		articlesSuccessToday:      3,
	}
	cache.Update(models.ServerStat{
		Total:           2, // should always replace existing total
		ArticlesTried:   5, // should be added to all existing values (historical gets "today" added, then "today" gets replaced)
		ArticlesSuccess: 5, // ditto ^
		DayParsed:       "2020-01-02",
	})
	require.Equal(2, cache.GetTotal())
	require.Equal(10, cache.GetArticlesTried())
	require.Equal(10, cache.GetArticlesSuccess())
}

func TestNewServerStatsCache_SetsServers(t *testing.T) {
	require := require.New(t)
	cache := NewServerStatsCache()
	require.NotNil(cache.Servers)
}

func TestUpdateServerStatsCache(t *testing.T) {
	require := require.New(t)
	cache := NewServerStatsCache()
	cache.Update(models.ServerStats{
		Total: 1,
		Servers: map[string]models.ServerStat{
			"server1": {
				Total:           1,
				ArticlesTried:   2,
				ArticlesSuccess: 2,
				DayParsed:       "2020-01-01",
			},
			"server2": {
				Total:           2,
				ArticlesTried:   4,
				ArticlesSuccess: 4,
				DayParsed:       "2020-01-01",
			},
		},
	})
	server1 := cache.Servers["server1"]
	server2 := cache.Servers["server2"]
	require.Equal(1, server1.GetTotal())
	require.Equal(2, server1.GetArticlesTried())
	require.Equal(2, server1.GetArticlesSuccess())
	require.Equal(2, server2.GetTotal())
	require.Equal(4, server2.GetArticlesTried())
	require.Equal(4, server2.GetArticlesSuccess())

	cache.Update(models.ServerStats{
		Total: 2,
		Servers: map[string]models.ServerStat{
			"server1": {
				Total:           3,
				ArticlesTried:   6,
				ArticlesSuccess: 6,
				DayParsed:       "2020-01-01",
			},
		},
	})
	server1 = cache.Servers["server1"]
	server2 = cache.Servers["server2"]
	require.Equal(2, cache.GetTotal())
	require.Equal(3, server1.GetTotal())
	require.Equal(6, server1.GetArticlesTried())
	require.Equal(6, server1.GetArticlesSuccess())
	require.Equal(2, server2.GetTotal())
	require.Equal(4, server2.GetArticlesTried())
	require.Equal(4, server2.GetArticlesSuccess())
}

func TestGetServerMap_ReturnsCopy(t *testing.T) {
	// It's important to return a true copy to maintain thread safety
	require := require.New(t)
	cache := NewServerStatsCache()
	cache.Update(models.ServerStats{
		Total: 1,
		Servers: map[string]models.ServerStat{
			"server1": {
				Total:           1,
				ArticlesTried:   2,
				ArticlesSuccess: 2,
				DayParsed:       "2020-01-01",
			},
		},
	})
	serverMap := cache.GetServerMap()
	require.Equal(cache.Servers, serverMap)
	require.NotSame(&cache.Servers, &serverMap)
	cache.Update(models.ServerStats{
		Total: 2,
		Servers: map[string]models.ServerStat{
			"server1": {
				Total:           3,
				ArticlesTried:   6,
				ArticlesSuccess: 6,
				DayParsed:       "2020-01-01",
			},
		},
	})
	cServer := cache.Servers["server1"]
	sServer := serverMap["server1"]
	require.NotEqual(cServer.GetTotal(), sServer.GetTotal())
	require.NotEqual(cServer.GetArticlesTried(), sServer.GetArticlesTried())
	require.NotEqual(cServer.GetArticlesSuccess(), sServer.GetArticlesSuccess())
}
