package main

import "microblog"

func main() {

	mapPostStore := microblog.MapPostStore{}
	mapPostStore.Post = map[string]string{}
	microblog.ListenAndServe(mapPostStore)

}
