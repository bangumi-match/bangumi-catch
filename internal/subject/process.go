package subject

func updateExistingFields(existing *JsonSubjectFile, newData *JsonSubjectFile) {
	existing.Collection = newData.Collection
	existing.Date = newData.Date
	existing.Eps = newData.Eps
	existing.Images = newData.Images
	existing.Infobox = newData.Infobox
	existing.Locked = newData.Locked
	existing.MetaTags = newData.MetaTags
	existing.Name = newData.Name
	existing.NameCn = newData.NameCn
	existing.Nsfw = newData.Nsfw
	existing.Platform = newData.Platform
	existing.Rating = newData.Rating
	existing.Series = newData.Series
	existing.Summary = newData.Summary
	existing.Tags = newData.Tags
	existing.TotalEpisodes = newData.TotalEpisodes
	existing.Type = newData.Type
	existing.Volumes = newData.Volumes
}
