package plugger 

import (
	"path/filepath"
	"io/ioutil"
)

func gatherPluginFiles(pluginFiles *[]string, pluginDirPath string, targetExt string) error {
    files, err := ioutil.ReadDir(pluginDirPath)
    if err != nil {
    	return err
    }
    for _, f := range files {
        fp := filepath.Join(pluginDirPath, f.Name())
        if f.IsDir() {
		if err := gatherPluginFiles(pluginFiles, fp, targetExt); err != nil {
			return err
		}
        } else {
		if filepath.Ext(f.Name()) == targetExt {
			absFp, err := filepath.Abs(fp)
			if err != nil {
				*pluginFiles = append(*pluginFiles, fp)
			} else {
				*pluginFiles = append(*pluginFiles, absFp)
			}
		}
        }
    }
    return nil
}
