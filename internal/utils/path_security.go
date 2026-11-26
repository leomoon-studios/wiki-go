package utils

import "wiki-go/internal/pathidentity"

// ResolvedDocumentPath pairs the identity used by access rules with its file path.
type ResolvedDocumentPath = pathidentity.ResolvedDocumentPath

func CanonicalRequestPath(value string) (string, error) {
	return pathidentity.CanonicalRequestPath(value)
}
func CanonicalDocumentRequestPath(value string) (string, error) {
	return pathidentity.CanonicalDocumentRequestPath(value)
}
func ResolveExistingDocumentPath(root, documents, value string) (ResolvedDocumentPath, error) {
	return pathidentity.ResolveExistingDocumentPath(root, documents, value)
}
func ResolveDocumentDestination(root, documents, value string) (ResolvedDocumentPath, error) {
	return pathidentity.ResolveDocumentDestination(root, documents, value)
}
func ResolveRelativeDocumentPath(root, documents, value string) (ResolvedDocumentPath, error) {
	return pathidentity.ResolveRelativeDocumentPath(root, documents, value)
}
func ResolveDocumentPath(root, documents, value string) (string, error) {
	return pathidentity.ResolveDocumentPath(root, documents, value)
}
func ResolveStoragePath(root, value string) (string, error) {
	return pathidentity.ResolveStoragePath(root, value)
}
