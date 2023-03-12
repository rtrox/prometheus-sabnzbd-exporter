package exporter

import (
	"prometheus-sabnzbd-exporter/internal/models"
	"sync"
)

type ServerStatCache struct {
	total                     int
	articlesTriedHistorical   int
	articlesTriedToday        int
	articlesSuccessHistorical int
	articlesSuccessToday      int
	todayKey                  string
}

func (s *ServerStatCache) Update(stat models.ServerStat) {
	s.total = stat.Total
	if stat.DayParsed != s.todayKey {
		s.articlesTriedHistorical += s.articlesTriedToday
		s.articlesSuccessHistorical += s.articlesSuccessToday
		s.articlesTriedToday = 0
		s.articlesSuccessToday = 0
		s.todayKey = stat.DayParsed
	}
	s.articlesTriedToday = stat.ArticlesTried
	s.articlesSuccessToday = stat.ArticlesSuccess
}

func (s *ServerStatCache) GetTotal() int {
	return s.total
}

func (s *ServerStatCache) GetArticlesTried() int {
	return s.articlesTriedHistorical + s.articlesTriedToday
}

func (s *ServerStatCache) GetArticlesSuccess() int {
	return s.articlesSuccessHistorical + s.articlesSuccessToday
}

type ServerStatsCache struct {
	lock    sync.RWMutex
	Total   int
	Servers map[string]ServerStatCache
}

func NewServerStatsCache() *ServerStatsCache {
	return &ServerStatsCache{
		Servers: make(map[string]ServerStatCache),
	}
}

func (c *ServerStatsCache) Update(stats models.ServerStats) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Total = stats.Total
	for name, srv := range stats.Servers {
		var toCache ServerStatCache
		if cached, ok := c.Servers[name]; ok {
			toCache = cached
		}
		toCache.Update(srv)
		c.Servers[name] = toCache
	}
}

func (c *ServerStatsCache) GetTotal() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.Total
}

func (c *ServerStatsCache) GetServerMap() map[string]ServerStatCache {
	c.lock.RLock()
	defer c.lock.RUnlock()
	ret := make(map[string]ServerStatCache)
	for k, v := range c.Servers {
		ret[k] = v
	}
	return ret
}
