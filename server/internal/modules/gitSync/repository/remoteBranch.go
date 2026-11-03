package repository

// remoteBranch возвращает имя удалённой рабочей ветки
func (r *GitSyncRepository) remoteBranch() string {
	return r.config.Remote + "/" + r.config.Branch
}
