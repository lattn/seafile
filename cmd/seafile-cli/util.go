package main

import "reflect"

func struct2map(in any) map[string]any {
	m := make(map[string]any)
	V := reflect.Indirect(reflect.ValueOf(in))
	T := V.Type()
	for i := 0; i < V.NumField(); i++ {
		m[T.Field(i).Name] = V.Field(i).Interface()
	}
	return m
}
