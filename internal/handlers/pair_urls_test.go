package handlers

import "testing"

func TestGenerateConsumerAdminURLs_DistinctPerCall(t *testing.T) {
	consumer, admin := GenerateConsumerAdminURLs()
	if consumer == admin {
		t.Fatal("consumer and admin should differ")
	}
}
