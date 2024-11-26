package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/rueidis"
	"github.com/spf13/cobra"
	"github.com/y7ut/potami/internal/conf"
)

var TrackerCmd = &cobra.Command{
	Use:   "tracker",
	Short: "Potami Tracker",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("please give a flow uuid")
			return
		}
		redis := conf.GetRedisClient()
		defer redis.Close()

		getTaskProgress := rueidis.NewLuaScript(`
			local results = {}
			for i, key in ipairs(KEYS) do
				if i > 4 then
					break
				end
				local progress_points = redis.call("SMEMBERS", "POTAMI:T:" .. key)
				
				-- 排序集合元素并获取最大值
				table.sort(progress_points, function(a, b) return tonumber(a) > tonumber(b) end)
				local max_progress = tonumber(progress_points[1]) or -1  -- 取排序后的第一个元素，即最大值

				-- 设置结果
				table.insert(results, {key, "" .. max_progress})
			end
			return results
		`)
		resp := getTaskProgress.Exec(context.Background(), redis, args, []string{})
		if resp.Error() != nil {
			log.Fatalf("failed to get flow progress: %v", resp.Error())
		}
		m, err := resp.ToArray()
		if err != nil {
			log.Fatalf("failed to get flow progress: %v", err)
		}

		for _, v := range m {
			info, _ := v.AsStrSlice()
			fmt.Printf("%s: %s\n", info[0], info[1])
		}
	},
}

func init() {
	RootCmd.AddCommand(TrackerCmd)
}
