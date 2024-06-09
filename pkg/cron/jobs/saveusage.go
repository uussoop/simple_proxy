package jobs

import (
	"strconv"

	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/pkg/cache"
)

func SaveUsage() {
	c := cache.GetCache()
	allusr, err := database.GetAllUsers()
	if err != nil {
		return
	}
	for _, u := range allusr {
		value, ok := c.Get(strconv.Itoa(int(u.ID)))
		if ok {
			databaselast, _ := c.Get(strconv.Itoa(int(u.ID)) + "cachedusage")
			if u.UsageToday != databaselast {
				c.Set(strconv.Itoa(int(u.ID)), u.UsageToday, 0)
				c.Set(strconv.Itoa(int(u.ID))+"cachedusage", u.UsageToday, 0)

			} else {

				database.UpdateUserUsageToday(u.ID, value.(int), false)
				c.Set(strconv.Itoa(int(u.ID))+"cachedusage", value.(int), 0)
			}

		}
	}

}
