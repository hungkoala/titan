package titan

func HealthCheck(subject string) (*Health, error) {
	client := GetDefaultClient()
	req, _ := NewReqBuilder().
		Get(SubjectToUrl(subject, "health")).
		Subject(subject).
		Build()
	var resp = &Health{}
	err := client.SendAndReceiveJson(NewBackgroundContext(), req, &resp)
	return resp, err
}
