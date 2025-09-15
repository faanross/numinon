package enumerate

import "github.com/faanross/numinon/internal/models"

type CommandEnumerate interface {
	DoEnumerate(args models.EnumerateArgs) (models.EnumerateResult, error)
}
