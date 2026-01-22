package resources

import (
	cr "github.com/alibabacloud-go/cr-20181201/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/arafato/ali-nuke/infrastructure"
	"github.com/arafato/ali-nuke/types"
	"github.com/arafato/ali-nuke/utils"
)

func init() {
	infrastructure.RegisterCollector("containerRegistryRepo", CollectContainerRegistryRepos)
}

// ContainerRegistryRepo represents an Alibaba Cloud Container Registry Repository
type ContainerRegistryRepo struct {
	Client     *cr.Client
	Region     string
	InstanceId string
}

// CollectContainerRegistryRepos discovers all Container Registry Repositories in the specified region
func CollectContainerRegistryRepos(creds *types.Credentials, region string) (types.Resources, error) {
	client, err := utils.CreateCRClient(creds, region)
	if err != nil {
		return nil, err
	}

	// First get all instances
	instanceResp, err := client.ListInstance(&cr.ListInstanceRequest{
		PageNo:   tea.Int32(1),
		PageSize: tea.Int32(100),
	})
	if err != nil {
		return nil, err
	}

	var allResources types.Resources

	if instanceResp.Body != nil && instanceResp.Body.Instances != nil {
		for _, instance := range instanceResp.Body.Instances {
			if instance.InstanceId == nil {
				continue
			}
			instanceID := *instance.InstanceId

			// Get repositories for this instance
			repoResp, err := client.ListRepository(&cr.ListRepositoryRequest{
				InstanceId: tea.String(instanceID),
				PageNo:     tea.Int32(1),
				PageSize:   tea.Int32(100),
			})
			if err != nil {
				continue
			}

			if repoResp.Body != nil && repoResp.Body.Repositories != nil {
				for _, repo := range repoResp.Body.Repositories {
					repoID := ""
					if repo.RepoId != nil {
						repoID = *repo.RepoId
					}

					repoName := ""
					if repo.RepoName != nil {
						repoName = *repo.RepoName
					}
					if repo.RepoNamespaceName != nil && *repo.RepoNamespaceName != "" {
						repoName = *repo.RepoNamespaceName + "/" + repoName
					}
					if repoName == "" {
						repoName = repoID
					}

					res := types.Resource{
						Removable:    ContainerRegistryRepo{Client: client, Region: region, InstanceId: instanceID},
						Region:       region,
						ResourceID:   repoID,
						ResourceName: repoName,
						ProductName:  "ContainerRegistryRepo",
					}
					allResources = append(allResources, &res)
				}
			}
		}
	}

	return allResources, nil
}

// Remove deletes the Container Registry Repository
func (c ContainerRegistryRepo) Remove(region string, resourceID string, resourceName string) error {
	request := &cr.DeleteRepositoryRequest{
		InstanceId: tea.String(c.InstanceId),
		RepoId:     tea.String(resourceID),
	}

	_, err := c.Client.DeleteRepository(request)
	return err
}
